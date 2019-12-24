package notify

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v2/utils"
	. "github.com/zouyx/agollo/v2/component/log"
	"strings"
	"sync"
	"time"
)

const (
	default_notification_id = -1
	comma = ","

	long_poll_interval        = 2 * time.Second //2s
)

var (
	allNotifications *notificationsMap
)

type notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int64  `json:"notificationId"`
}

type notificationsMap struct {
	notifications map[string]int64
	sync.RWMutex
}

type apolloNotify struct {
	NotificationId int64  `json:"notificationId"`
	NamespaceName  string `json:"namespaceName"`
}

func (n *notificationsMap) setNotify(namespaceName string, notificationId int64) {
	n.Lock()
	defer n.Unlock()
	n.notifications[namespaceName] = notificationId
}

func (n *notificationsMap) getNotify(namespace string) int64 {
	n.RLock()
	defer n.RUnlock()
	return n.notifications[namespace]
}

func (n *notificationsMap) getNotifies(namespace string) string {
	n.RLock()
	defer n.RUnlock()

	notificationArr := make([]*notification, 0)
	if namespace == "" {
		for namespaceName, notificationId := range n.notifications {
			notificationArr = append(notificationArr,
				&notification{
					NamespaceName:  namespaceName,
					NotificationId: notificationId,
				})
		}
	} else {
		n := n.notifications[namespace]
		notificationArr = append(notificationArr,
			&notification{
				NamespaceName:  namespace,
				NotificationId: n,
			})
	}

	j, err := json.Marshal(notificationArr)

	if err != nil {
		return ""
	}

	return string(j)
}

func initAllNotifications() {
	if appConfig == nil {
		allNotifications = &notificationsMap{
			notifications: make(map[string]int64, 0),
		}
		return
	}
	namespaces := SplitNamespaces(appConfig.NamespaceName,
		func(namespace string) {})

	allNotifications = &notificationsMap{
		notifications: namespaces,
	}
}

type NotifyConfigComponent struct {
}

func (this *NotifyConfigComponent) Start() {
	t2 := time.NewTimer(long_poll_interval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			notifySyncConfigServices()
			t2.Reset(long_poll_interval)
		}
	}
}

func notifySyncConfigServices() error {

	remoteConfigs, err := notifyRemoteConfig(nil, utils.Empty)

	if err != nil {
		return fmt.Errorf("notifySyncConfigServices: %s", err)
	}
	if len(remoteConfigs) == 0 {
		return fmt.Errorf("notifySyncConfigServices: empty remote config")
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	err = autoSyncConfigServices(nil)

	//first sync fail then load config file
	if err != nil {
		SplitNamespaces(appConfig.NamespaceName, func(namespace string) {
			config, _ := loadConfigFile(appConfig.BackupConfigPath, namespace)
			if config != nil {
				updateApolloConfig(config, false)
			}
		})
	}
	//sync all config
	return nil
}

func notifySimpleSyncConfigServices(namespace string) error {

	remoteConfigs, err := notifyRemoteConfig(nil, namespace)

	if err != nil || len(remoteConfigs) == 0 {
		return err
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	notifications := make(map[string]int64)
	notifications[remoteConfigs[0].NamespaceName] = remoteConfigs[0].NotificationId

	return autoSyncNamespaceConfigServices(nil, notifications)
}

func toApolloConfig(resBody []byte) ([]*apolloNotify, error) {
	remoteConfig := make([]*apolloNotify, 0)

	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		Logger.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	return remoteConfig, nil
}

func notifyRemoteConfig(newAppConfig *AppConfig, namespace string) ([]*apolloNotify, error) {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyUrlSuffix(allNotifications.getNotifies(namespace), appConfig, newAppConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	notifies, err := requestRecovery(appConfig, &ConnectConfig{
		Uri:     urlSuffix,
		Timeout: nofity_connect_timeout,
	}, &CallBack{
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

func updateAllNotifications(remoteConfigs []*apolloNotify) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}
		if allNotifications.getNotify(remoteConfig.NamespaceName) == 0 {
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationId)
	}
}

func autoSyncConfigServicesSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	apolloConfig, err := createApolloConfigWithJson(responseBody)

	if err != nil {
		Logger.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}

	updateApolloConfig(apolloConfig, appConfig.getIsBackupConfig())

	return nil, nil
}

func autoSyncConfigServices(newAppConfig *AppConfig) error {
	return autoSyncNamespaceConfigServices(newAppConfig, allNotifications.notifications)
}

func autoSyncNamespaceConfigServices(newAppConfig *AppConfig, notifications map[string]int64) error {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	var err error
	for namespace := range notifications {
		urlSuffix := getConfigURLSuffix(appConfig, namespace)

		_, err = requestRecovery(appConfig, &ConnectConfig{
			Uri: urlSuffix,
		}, &CallBack{
			SuccessCallBack:   autoSyncConfigServicesSuccessCallBack,
			NotModifyCallBack: touchApolloConfigCache,
		})
		if err != nil {
			return err
		}
	}
	return err
}

func SplitNamespaces(namespacesStr string, callback func(namespace string)) map[string]int64 {
	namespaces := make(map[string]int64, 1)
	split := strings.Split(namespacesStr, comma)
	for _, namespace := range split {
		callback(namespace)
		namespaces[namespace] = default_notification_id
	}
	return namespaces
}
