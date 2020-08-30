/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package notify

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v4/constant"
	"net/url"
	"path"
	"time"

	"github.com/zouyx/agollo/v4/component"
	"github.com/zouyx/agollo/v4/component/log"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/extension"
	"github.com/zouyx/agollo/v4/protocol/http"
	"github.com/zouyx/agollo/v4/storage"
	"github.com/zouyx/agollo/v4/utils"
)

const (
	longPollInterval = 2 * time.Second //2s

	//notify timeout
	nofityConnectTimeout = 10 * time.Minute //10m

	//同步链接时间
	syncNofityConnectTimeout = 3 * time.Second //3s

	defaultContentKey = "content"
)

//ConfigComponent 配置组件
type ConfigComponent struct {
	appConfig *config.AppConfig
}

//Start 启动配置组件定时器
func (c *ConfigComponent) Start() {
	t2 := time.NewTimer(longPollInterval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			AsyncConfigs(c.appConfig)
			t2.Reset(longPollInterval)
		}
	}
}

//AsyncConfigs 异步同步所有配置文件中配置的namespace配置
func AsyncConfigs(appConfig *config.AppConfig) error {
	return syncConfigs(utils.Empty, true, appConfig)
}

//SyncConfigs 同步同步所有配置文件中配置的namespace配置
func SyncConfigs(appConfig *config.AppConfig) error {
	return syncConfigs(utils.Empty, false, appConfig)
}

//SyncNamespaceConfig 同步同步一个指定的namespace配置
func SyncNamespaceConfig(namespace string, appConfig *config.AppConfig) error {
	return syncConfigs(namespace, false, appConfig)
}

func syncConfigs(namespace string, isAsync bool, appConfig *config.AppConfig) error {

	remoteConfigs, err := notifyRemoteConfig(nil, namespace, isAsync)

	if err != nil || len(remoteConfigs) == 0 {
		loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	if err != nil {
		return fmt.Errorf("notifySyncConfigServices: %s", err)
	}
	if len(remoteConfigs) == 0 {
		return fmt.Errorf("notifySyncConfigServices: empty remote config")
	}

	appConfig.GetNotificationsMap().UpdateAllNotifications(remoteConfigs)

	//sync all config
	err = AutoSyncConfigServices(nil)

	if err != nil {
		if namespace != "" {
			return nil
		}
		//first sync fail then load config file
		loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	//sync all config
	return nil
}

func loadBackupConfig(namespace string, appConfig *config.AppConfig) {
	env.SplitNamespaces(namespace, func(namespace string) {
		config, _ := extension.GetFileHandler().LoadConfigFile(appConfig.BackupConfigPath, namespace)
		if config != nil {
			storage.UpdateApolloConfig(config, false)
		}
	})
}

func toApolloConfig(resBody []byte) ([]*config.Notification, error) {
	remoteConfig := make([]*config.Notification, 0)

	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		log.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	return remoteConfig, nil
}

func notifyRemoteConfig(appConfig *config.AppConfig, namespace string, isAsync bool) ([]*config.Notification, error) {
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyURLSuffix(appConfig.GetNotificationsMap().GetNotifies(namespace), appConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	connectConfig := &env.ConnectConfig{
		URI:    urlSuffix,
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if !isAsync {
		connectConfig.Timeout = syncNofityConnectTimeout
	} else {
		connectConfig.Timeout = nofityConnectTimeout
	}
	connectConfig.IsRetry = isAsync
	notifies, err := http.RequestRecovery(appConfig, connectConfig, &http.CallBack{
		SuccessCallBack: func(responseBody []byte) (interface{}, error) {
			return toApolloConfig(responseBody)
		},
		NotModifyCallBack: touchApolloConfigCache,
	})

	if notifies == nil {
		return nil, err
	}

	return notifies.([]*config.Notification), err
}
func touchApolloConfigCache() error {
	return nil
}

//AutoSyncConfigServicesSuccessCallBack 同步配置回调
func AutoSyncConfigServicesSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	apolloConfig, err := createApolloConfigWithJSON(responseBody)

	if err != nil {
		log.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	appConfig := env.GetPlainAppConfig()

	storage.UpdateApolloConfig(apolloConfig, appConfig.GetIsBackupConfig())

	return nil, nil
}

// createApolloConfigWithJSON 使用json配置转换成apolloconfig
func createApolloConfigWithJSON(b []byte) (*env.ApolloConfig, error) {
	apolloConfig := &env.ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	parser := extension.GetFormatParser(constant.ConfigFileFormat(path.Ext(apolloConfig.NamespaceName)))
	if parser == nil {
		parser = extension.GetFormatParser(constant.DEFAULT)
	}

	if parser == nil {
		return apolloConfig, nil
	}
	m, err := parser.Parse(apolloConfig.Configurations[defaultContentKey])
	if err != nil {
		log.Debug("GetContent fail ! error:", err)
	}

	if len(m) > 0 {
		apolloConfig.Configurations = m
	}
	return apolloConfig, nil
}

//AutoSyncConfigServices 自动同步配置
func AutoSyncConfigServices(newAppConfig *config.AppConfig) error {
	return autoSyncNamespaceConfigServices(newAppConfig)
}

func autoSyncNamespaceConfigServices(appConfig *config.AppConfig) error {
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	var err error
	appConfig.GetNotificationsMap().GetNotifications().Range(func(key, value interface{}) bool {
		namespace := key.(string)
		urlSuffix := component.GetConfigURLSuffix(appConfig, namespace)

		_, err = http.RequestRecovery(appConfig, &env.ConnectConfig{
			URI:    urlSuffix,
			AppID:  appConfig.AppID,
			Secret: appConfig.Secret,
		}, &http.CallBack{
			SuccessCallBack:   AutoSyncConfigServicesSuccessCallBack,
			NotModifyCallBack: touchApolloConfigCache,
		})
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func getNotifyURLSuffix(notifications string, config *config.AppConfig) string {
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}
