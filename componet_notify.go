package agollo

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/cihub/seelog"
)

const (
	default_notification_id = -1
)

var (
	allNotifications *notificationsMap
)

type NotifyConfigComponent struct {
}

type apolloNotify struct {
	NotificationId int64  `json:"notificationId"`
	NamespaceName  string `json:"namespaceName"`
}

func (this *NotifyConfigComponent) Start() {
	//long poll for sync
	for {
		notifySyncConfigServices()
		time.Sleep(time.Second/2)
	}
}

func toApolloConfig(resBody []byte) ([]*apolloNotify, error) {
	remoteConfig := make([]*apolloNotify, 0)

	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		seelog.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	return remoteConfig, nil
}

func getRemoteConfigSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	return toApolloConfig(responseBody)
}

func getRemoteConfig() ([]*apolloNotify, error) {
	appConfig := GetAppConfig()
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyUrlSuffix(allNotifications.getNotifies(), appConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	notifies, err := requestRecovery(appConfig, urlSuffix, getRemoteConfigSuccessCallBack)

	if notifies == nil {
		return nil, err
	}

	return notifies.([]*apolloNotify), err
}

func notifySyncConfigServices() error {

	remoteConfigs, err := getRemoteConfig()

	if err != nil || len(remoteConfigs) == 0 {
		return err
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	SyncConfig()

	return nil
}

func updateAllNotifications(remoteConfigs []*apolloNotify) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationId)
	}
}

func initNotify() {
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}
	appConfig := GetAppConfig()

	allNotifications.setNotify(appConfig.NamespaceName, default_notification_id)
}

type notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int64  `json:"notificationId"`
}

type notificationsMap struct {
	notifications map[string]int64
	sync.RWMutex
}

func (this *notificationsMap) setNotify(namespaceName string, notificationId int64) {
	this.Lock()
	defer this.Unlock()
	this.notifications[namespaceName] = notificationId
}
func (this *notificationsMap) getNotifies() string {
	this.RLock()
	defer this.RUnlock()

	notificationArr := make([]*notification, 0)
	for namespaceName, notificationId := range this.notifications {
		notificationArr = append(notificationArr,
			&notification{
				NamespaceName:  namespaceName,
				NotificationId: notificationId,
			})
	}

	j, err := json.Marshal(notificationArr)

	if err != nil {
		return ""
	}

	return string(j)
}
