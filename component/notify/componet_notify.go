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
	"github.com/zouyx/agollo/v3/constant"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/zouyx/agollo/v3/component"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/protocol/http"
	"github.com/zouyx/agollo/v3/storage"
	"github.com/zouyx/agollo/v3/utils"
)

const (
	longPollInterval = 2 * time.Second //2s

	//notify timeout
	nofityConnectTimeout = 10 * time.Minute //10m

	//同步链接时间
	syncNofityConnectTimeout = 3 * time.Second //3s

	defaultNotificationId = int64(-1)

	defaultContentKey = "content"
)

var (
	allNotifications *notificationsMap
)

type notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationID int64  `json:"notificationId"`
}

// map[string]int64
type notificationsMap struct {
	notifications sync.Map
}

type apolloNotify struct {
	NotificationID int64  `json:"notificationId"`
	NamespaceName  string `json:"namespaceName"`
}

//InitAllNotifications 初始化notificationsMap
func InitAllNotifications(callback func(namespace string)) {
	appConfig := env.GetPlainAppConfig()
	ns := env.SplitNamespaces(appConfig.NamespaceName, callback)
	allNotifications = &notificationsMap{
		notifications: ns,
	}
}

func (n *notificationsMap) setNotify(namespaceName string, notificationID int64) {
	n.notifications.Store(namespaceName, notificationID)
}

func (n *notificationsMap) getNotify(namespace string) int64 {
	value, ok := n.notifications.Load(namespace)
	if !ok || value == nil {
		return 0
	}
	return value.(int64)
}

func (n *notificationsMap) GetNotifyLen() int {
	s := n.notifications
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

func (n *notificationsMap) getNotifies(namespace string) string {
	notificationArr := make([]*notification, 0)
	if namespace == "" {
		n.notifications.Range(func(key, value interface{}) bool {
			namespaceName := key.(string)
			notificationID := value.(int64)
			notificationArr = append(notificationArr,
				&notification{
					NamespaceName:  namespaceName,
					NotificationID: notificationID,
				})
			return true
		})
	} else {
		notify, _ := n.notifications.LoadOrStore(namespace, defaultNotificationId)

		notificationArr = append(notificationArr,
			&notification{
				NamespaceName:  namespace,
				NotificationID: notify.(int64),
			})
	}

	j, err := json.Marshal(notificationArr)

	if err != nil {
		return ""
	}

	return string(j)
}

//ConfigComponent 配置组件
type ConfigComponent struct {
}

//Start 启动配置组件定时器
func (c *ConfigComponent) Start() {
	t2 := time.NewTimer(longPollInterval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			AsyncConfigs()
			t2.Reset(longPollInterval)
		}
	}
}

//AsyncConfigs 异步同步所有配置文件中配置的namespace配置
func AsyncConfigs() error {
	return syncConfigs(utils.Empty, true)
}

//SyncConfigs 同步同步所有配置文件中配置的namespace配置
func SyncConfigs() error {
	return syncConfigs(utils.Empty, false)
}

//SyncNamespaceConfig 同步同步一个指定的namespace配置
func SyncNamespaceConfig(namespace string) error {
	return syncConfigs(namespace, false)
}

func syncConfigs(namespace string, isAsync bool) error {

	remoteConfigs, err := notifyRemoteConfig(nil, namespace, isAsync)

	if err != nil || len(remoteConfigs) == 0 {
		appConfig := env.GetPlainAppConfig()
		loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	if err != nil {
		return fmt.Errorf("notifySyncConfigServices: %s", err)
	}
	if len(remoteConfigs) == 0 {
		return fmt.Errorf("notifySyncConfigServices: empty remote config")
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	err = AutoSyncConfigServices(nil)

	if err != nil {
		if namespace != "" {
			return nil
		}
		//first sync fail then load config file
		appConfig := env.GetPlainAppConfig()
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

func toApolloConfig(resBody []byte) ([]*apolloNotify, error) {
	remoteConfig := make([]*apolloNotify, 0)

	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		log.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	return remoteConfig, nil
}

func notifyRemoteConfig(newAppConfig *config.AppConfig, namespace string, isAsync bool) ([]*apolloNotify, error) {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyURLSuffix(allNotifications.getNotifies(namespace), appConfig, newAppConfig)

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

	return notifies.([]*apolloNotify), err
}
func touchApolloConfigCache() error {
	return nil
}

func updateAllNotifications(remoteConfigs []*apolloNotify) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}
		if allNotifications.getNotify(remoteConfig.NamespaceName) == 0 {
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationID)
	}
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
	return autoSyncNamespaceConfigServices(newAppConfig, allNotifications)
}

func autoSyncNamespaceConfigServices(newAppConfig *config.AppConfig, allNotifications *notificationsMap) error {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	var err error
	allNotifications.notifications.Range(func(key, value interface{}) bool {
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

func getNotifyURLSuffix(notifications string, config *config.AppConfig, newConfig *config.AppConfig) string {
	c := config
	if newConfig != nil {
		c = newConfig
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(c.AppID),
		url.QueryEscape(c.Cluster),
		url.QueryEscape(notifications))
}
