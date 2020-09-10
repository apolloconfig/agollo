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
	"github.com/zouyx/agollo/v4/storage"
	"net/url"
	"path"
	"time"

	"github.com/zouyx/agollo/v4/component"
	"github.com/zouyx/agollo/v4/component/log"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/extension"
	"github.com/zouyx/agollo/v4/protocol/http"
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
	cache     *storage.Cache
}

// SetAppConfig nolint
func (c *ConfigComponent) SetAppConfig(appConfig *config.AppConfig) {
	c.appConfig = appConfig
}

// SetCache nolint
func (c *ConfigComponent) SetCache(cache *storage.Cache) {
	c.cache = cache
}

//Start 启动配置组件定时器
func (c *ConfigComponent) Start() {
	t2 := time.NewTimer(longPollInterval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			configs := AsyncConfigs(c.appConfig)
			for _, apolloConfig := range configs {
				c.cache.UpdateApolloConfig(apolloConfig, c.appConfig, true)
			}
			t2.Reset(longPollInterval)
		}
	}
}

//AsyncConfigs 异步同步所有配置文件中配置的namespace配置
func AsyncConfigs(appConfig *config.AppConfig) []*config.ApolloConfig {
	remoteConfigs, err := notifyRemoteConfig(appConfig, utils.Empty)

	var apolloConfigs []*config.ApolloConfig
	if err != nil || len(remoteConfigs) == 0 {
		apolloConfigs = loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	if len(apolloConfigs) > 0 {
		return apolloConfigs
	}

	appConfig.GetNotificationsMap().UpdateAllNotifications(remoteConfigs)

	notifications := appConfig.GetNotificationsMap().GetNotifications()
	n := &notifications
	n.Range(func(key, value interface{}) bool {
		apolloConfig := syncConfigs(utils.Empty, component.GetConfigURLSuffix, appConfig)
		apolloConfigs = append(apolloConfigs, apolloConfig)
		return true
	})
	return apolloConfigs

}

//SyncConfigs 同步同步所有配置文件中配置的namespace配置
func SyncConfigs(appConfig *config.AppConfig) *config.ApolloConfig {
	return SyncNamespaceConfig(utils.Empty, appConfig)
}

//SyncNamespaceConfig 同步同步一个指定的namespace配置
func SyncNamespaceConfig(namespace string, appConfig *config.AppConfig) *config.ApolloConfig {
	return syncConfigs(namespace, component.GetConfigFilesURLSuffix, appConfig)
}

func syncConfigs(namespace string, getURI func(config *config.AppConfig, namespaceName string) string, appConfig *config.AppConfig) *config.ApolloConfig {
	//sync all config
	return autoSyncNamespaceConfigServices(namespace, appConfig, getURI)
}

func loadBackupConfig(namespace string, appConfig *config.AppConfig) []*config.ApolloConfig {
	apolloConfigs := make([]*config.ApolloConfig, 0)
	config.SplitNamespaces(namespace, func(namespace string) {
		c, err := extension.GetFileHandler().LoadConfigFile(appConfig.BackupConfigPath, namespace)
		if err != nil {
			log.Error("LoadConfigFile error, error", err)
			return
		}
		if c == nil {
			return
		}
		apolloConfigs = append(apolloConfigs, c)
	})
	return apolloConfigs
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

func notifyRemoteConfig(appConfig *config.AppConfig, namespace string) ([]*config.Notification, error) {
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyURLSuffix(appConfig.GetNotificationsMap().GetNotifies(namespace), appConfig)

	connectConfig := &env.ConnectConfig{
		URI:    urlSuffix,
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	connectConfig.Timeout = nofityConnectTimeout
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
	return createApolloConfigWithJSON(responseBody)
}

// createApolloConfigWithJSON 使用json配置转换成apolloconfig
func createApolloConfigWithJSON(b []byte) (*config.ApolloConfig, error) {
	apolloConfig := &config.ApolloConfig{}
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
func AutoSyncConfigServices(newAppConfig *config.AppConfig) []*config.ApolloConfig {
	configs := make([]*config.ApolloConfig, 0, 8)
	config.SplitNamespaces(newAppConfig.NamespaceName, func(namespace string) {
		apolloConfig := autoSyncNamespaceConfigServices(namespace, newAppConfig, component.GetConfigFilesURLSuffix)
		if apolloConfig == nil {
			return
		}

		configs = append(configs, apolloConfig)
	})
	return configs
}

func autoSyncNamespaceConfigServices(namespace string, appConfig *config.AppConfig, getURI func(config *config.AppConfig, namespaceName string) string) *config.ApolloConfig {
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	urlSuffix := getURI(appConfig, namespace)

	apolloConfig, err := http.RequestRecovery(appConfig, &env.ConnectConfig{
		URI:    urlSuffix,
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}, &http.CallBack{
		SuccessCallBack:   AutoSyncConfigServicesSuccessCallBack,
		NotModifyCallBack: touchApolloConfigCache,
	})
	if err != nil {
		log.Errorf("request %s fail, error:%v", urlSuffix, err)
		return nil
	}

	if apolloConfig == nil {
		log.Warn("apolloConfig is nil")
		return nil
	}

	return apolloConfig.(*config.ApolloConfig)
}

func getNotifyURLSuffix(notifications string, config *config.AppConfig) string {
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}
