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

package config

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v4/component/log"
	"github.com/zouyx/agollo/v4/utils"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	//next try connect period - 60 second
	nextTryConnectPeriod int64 = 60

	defaultNotificationID = int64(-1)
	comma                 = ","
)

//File 读写配置文件
type File interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

//AppConfig 配置文件
type AppConfig struct {
	AppID            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	IP               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
	Secret           string `json:"secret"`
	//real servers ip
	servers                 sync.Map
	notificationsMap        *notificationsMap
	currentConnApolloConfig *CurrentApolloConfig
}

//ServerInfo 服务器信息
type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

//GetIsBackupConfig whether backup config after fetch config from apollo
//false : no
//true : yes (default)
func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

//GetBackupConfigPath GetBackupConfigPath
func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
}

//GetHost GetHost
func (a *AppConfig) GetHost() string {
	if strings.HasPrefix(a.IP, "http") {
		if !strings.HasSuffix(a.IP, "/") {
			return a.IP + "/"
		}
		return a.IP
	}
	return "http://" + a.IP + "/"
}

// InitAllNotifications 初始化notificationsMap
func (a *AppConfig) Init() {
	a.currentConnApolloConfig = CreateCurrentApolloConfig()
}

type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationID int64  `json:"notificationId"`
}

// InitAllNotifications 初始化notificationsMap
func (a *AppConfig) InitAllNotifications(callback func(namespace string)) {
	ns := SplitNamespaces(a.NamespaceName, callback)
	a.notificationsMap = &notificationsMap{
		notifications: ns,
	}
}

//SplitNamespaces 根据namespace字符串分割后，并执行callback函数
func SplitNamespaces(namespacesStr string, callback func(namespace string)) sync.Map {
	namespaces := sync.Map{}
	split := strings.Split(namespacesStr, comma)
	for _, namespace := range split {
		if callback != nil {
			callback(namespace)
		}
		namespaces.Store(namespace, defaultNotificationID)
	}
	return namespaces
}

// GetNotifications 获取notificationsMap
func (a *AppConfig) GetNotificationsMap() *notificationsMap {
	return a.notificationsMap
}

//SetNextTryConnTime if this connect is fail will set this time
func (a *AppConfig) SetNextTryConnTime(nextTryConnectPeriod int64) {
	a.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//IsConnectDirectly is connect by ip directly
//false : no
//true : yes
func (a *AppConfig) IsConnectDirectly() bool {
	if a.NextTryConnTime >= 0 && a.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

//SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func (a *AppConfig) SyncServerIPListSuccessCallBack(appConfig *AppConfig, responseBody []byte) (o interface{}, err error) {
	log.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*ServerInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		log.Error("Unmarshal json Fail,Error:", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		a.servers.Store(server.HomepageURL, server)
	}
	return
}

//SetDownNode 设置失效节点
func (a *AppConfig) SetDownNode(host string) {
	if host == "" {
		return
	}

	if host == a.GetHost() {
		a.SetNextTryConnTime(nextTryConnectPeriod)
	}

	a.GetServers().Range(func(k, v interface{}) bool {
		server := v.(*ServerInfo)
		// if some node has down then select next node
		if strings.Index(k.(string), host) > -1 {
			server.IsDown = true
			return false
		}
		return true
	})
}

//GetServers 获取服务器数组
func (a *AppConfig) GetServers() *sync.Map {
	return &a.servers
}

//GetServersLen 获取服务器数组长度
func (a *AppConfig) GetServersLen() int {
	s := a.GetServers()
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

//GetServicesConfigURL 获取服务器列表url
func (a *AppConfig) GetServicesConfigURL() string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		a.GetHost(),
		url.QueryEscape(a.AppID),
		utils.GetInternal())
}

// nolint
func (a *AppConfig) SetCurrentApolloConfig(apolloConfig *ApolloConnConfig) {
	a.currentConnApolloConfig.Set(apolloConfig.NamespaceName, apolloConfig)
}

// nolint
func (a *AppConfig) GetCurrentApolloConfig() *CurrentApolloConfig {
	return a.currentConnApolloConfig
}

// map[string]int64
type notificationsMap struct {
	notifications sync.Map
}

func (n *notificationsMap) UpdateAllNotifications(remoteConfigs []*Notification) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}
		if n.GetNotify(remoteConfig.NamespaceName) == 0 {
			continue
		}

		n.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationID)
	}
}

func (n *notificationsMap) setNotify(namespaceName string, notificationID int64) {
	n.notifications.Store(namespaceName, notificationID)
}

func (n *notificationsMap) GetNotify(namespace string) int64 {
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

func (n *notificationsMap) GetNotifications() sync.Map {
	return n.notifications
}

func (n *notificationsMap) GetNotifies(namespace string) string {
	notificationArr := make([]*Notification, 0)
	if namespace == "" {
		n.notifications.Range(func(key, value interface{}) bool {
			namespaceName := key.(string)
			notificationID := value.(int64)
			notificationArr = append(notificationArr,
				&Notification{
					NamespaceName:  namespaceName,
					NotificationID: notificationID,
				})
			return true
		})
	} else {
		notify, _ := n.notifications.LoadOrStore(namespace, defaultNotificationID)

		notificationArr = append(notificationArr,
			&Notification{
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
