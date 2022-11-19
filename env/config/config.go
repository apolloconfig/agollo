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
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/qshuai/agollo/v4/utils"
)

var (
	defaultNotificationID = int64(-1)
	comma                 = ","
	randSource            = rand.NewSource(time.Now().UnixNano())

	namespaceFetcher    NamespaceFetcher
	adminServiceFetcher AdminServiceFetcher
)

// File 读写配置文件
type File interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

type NamespaceFetcher interface {
	Fetch(config *AppConfig) ([]string, error)
}

type AdminServiceFetcher interface {
	Fetch(config *AppConfig) ([]string, error)
}

// AppConfig 配置文件
type AppConfig struct {
	AppID                   string        `json:"appId"`
	Cluster                 string        `json:"cluster"`
	NamespaceName           string        `json:"namespaceName"`
	Dynamic                 bool          `json:"dynamic"`                 // 是否开启正则表达式形式的namespace配置
	SyncNamespaceInterval   time.Duration `json:"sync_namespace_interval"` // 拉取最新namespace的时间间隔
	IP                      string        `json:"ip"`
	IsBackupConfig          bool          `default:"true" json:"isBackupConfig"`
	BackupConfigPath        string        `json:"backupConfigPath"`
	Secret                  string        `json:"secret"`
	Label                   string        `json:"label"`
	SyncServerTimeout       int           `json:"syncServerTimeout"`   // 与apollo server端同步数据的超时时间设置，单位：秒
	MustStart               bool          `default:"false"`            // MustStart 可用于控制第一次同步必须成功
	AuthorizationToken      string        `json:"authorization_token"` // 自定义的认证
	notificationsMap        *notificationsMap
	currentConnApolloConfig *CurrentApolloConfig

	*sync.RWMutex                        // protecting the following fields
	namespace        string              // the real namespace list
	pattern          []*regexp.Regexp    // 如果Dynamic设置为true，那么该字段存储正则列表；否则为nil
	namespaceMapping map[string]struct{} // 类型为map[string]struct{}；保存namespace的集合，key: 单个namespace value: struct{}{}; 便于快速过滤namespace
}

// ServerInfo 服务器信息
type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

// GetNamespace 获取最新的namespace
func (a *AppConfig) GetNamespace() string {
	a.RLock()
	ns := a.namespace
	a.RUnlock()
	return ns
}

// AddNamespace 添加新的namespace
func (a *AppConfig) AddNamespace(namespace string) {
	a.Lock()
	defer a.Unlock()

	SplitNamespaces(namespace, func(namespace string) {
		if _, ok := a.namespaceMapping[namespace]; ok {
			return
		}

		a.namespaceMapping[namespace] = struct{}{}

		if a.namespace == "" {
			a.namespace = namespace
		} else {
			a.namespace += ("," + namespace)
		}
		a.notificationsMap.notifications.Store(namespace, defaultNotificationID)
	})
}

// IsFresh 判断namespace是否为新添加的
func (a *AppConfig) IsFresh(namespace string) bool {
	a.RLock()
	_, ok := a.namespaceMapping[namespace]
	a.RUnlock()
	if ok {
		return false
	}

	for _, x := range a.pattern {
		if x.MatchString(namespace) {
			return true
		}
	}

	return false
}

// GetIsBackupConfig whether backup config after fetch config from apollo
// false : no
// true : yes (default)
func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

// GetBackupConfigPath GetBackupConfigPath
func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
}

// GetHost GetHost
func (a *AppConfig) GetHost() string {
	u, err := url.Parse(a.IP)
	if err != nil {
		return a.IP
	}
	if !strings.HasSuffix(u.Path, "/") {
		return u.String() + "/"
	}
	return u.String()
}

// Init 初始化notificationsMap
func (a *AppConfig) Init() {
	a.RWMutex = &sync.RWMutex{}

	// 设置real namespace
	a.namespace = a.NamespaceName
	if a.Dynamic {
		// 解析正则
		var regexList []*regexp.Regexp
		SplitNamespaces(a.NamespaceName, func(namespace string) {
			regexList = append(regexList, regexp.MustCompile(namespace))
		})
		a.pattern = regexList

		allNamespaces, err := namespaceFetcher.Fetch(a)
		if err != nil {
			panic("sync apollo namespace err: " + err.Error())
		}
		if len(allNamespaces) == 0 {
			panic("apollo sever's namespace not config")
		}

		var targetNs []string
		for _, ns := range allNamespaces {
			for _, x := range a.pattern {
				if x.MatchString(ns) {
					targetNs = append(targetNs, ns)
				}
			}
		}
		if len(targetNs) == 0 {
			panic("no matched namespace, please check config: " + a.NamespaceName)
		}

		a.namespace = strings.Join(targetNs, comma)
	}

	// 初始化内部字段
	a.namespaceMapping = make(map[string]struct{})
	SplitNamespaces(a.namespace, func(namespace string) {
		a.namespaceMapping[namespace] = struct{}{}
	})

	a.currentConnApolloConfig = CreateCurrentApolloConfig()
	a.initAllNotifications(nil)
}

// Notification 用于保存 apollo Notification 信息
type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationID int64  `json:"notificationId"`
}

// InitAllNotifications 初始化notificationsMap
func (a *AppConfig) initAllNotifications(callback func(namespace string)) {
	ns := SplitNamespaces(a.GetNamespace(), callback)
	a.notificationsMap = &notificationsMap{
		notifications: ns,
	}
}

// SplitNamespaces 根据namespace字符串分割后，并执行callback函数
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

// JoinNamespace 将多个namespace按照分割符进行连接
func JoinNamespace(namespaces []string) string {
	if len(namespaces) == 0 {
		return ""
	}

	return strings.Join(namespaces, comma)
}

// GetNotificationsMap 获取notificationsMap
func (a *AppConfig) GetNotificationsMap() *notificationsMap {
	return a.notificationsMap
}

// GetServicesConfigURL 获取服务器列表url
func (a *AppConfig) GetServicesConfigURL() string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		a.GetHost(),
		url.QueryEscape(a.AppID),
		utils.GetInternal())
}

// GetAdminServiceURL 获取admin service列表
func (a *AppConfig) GetAdminServiceURL() string {
	return fmt.Sprintf("%sservices/admin", a.GetHost())
}

// GetNamespaceListURL 获取appID的namespace列表url
func (a *AppConfig) GetNamespaceListURL() (string, error) {
	services, err := adminServiceFetcher.Fetch(a)
	if err != nil {
		return "", err
	}
	if len(services) == 0 {
		return "", errors.New("admin service instance is empty")
	}

	return fmt.Sprintf("%sapps/%s/clusters/%s/namespaces", services[int(randSource.Int63())%len(services)], a.AppID, a.Cluster), nil
}

// SetCurrentApolloConfig nolint
func (a *AppConfig) SetCurrentApolloConfig(apolloConfig *ApolloConnConfig) {
	a.currentConnApolloConfig.Set(apolloConfig.NamespaceName, apolloConfig)
}

// GetCurrentApolloConfig nolint
func (a *AppConfig) GetCurrentApolloConfig() *CurrentApolloConfig {
	return a.currentConnApolloConfig
}

type notificationsMap struct {
	// type: map[string]int64
	// business: map[namespace]notificationID
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

// UpdateNotify update namespace's notification ID
func (n *notificationsMap) UpdateNotify(namespaceName string, notificationID int64) {
	if namespaceName != "" {
		n.setNotify(namespaceName, notificationID)
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

func SetAdminServiceFetcher(fetcher AdminServiceFetcher) {
	adminServiceFetcher = fetcher
}

func SetNamespaceFetcher(fetcher NamespaceFetcher) {
	namespaceFetcher = fetcher
}
