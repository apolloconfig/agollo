package notify

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v2/env/config"
	"net/url"
	"sync"
	"time"

	"github.com/zouyx/agollo/v2/component"
	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/protocol/http"
	"github.com/zouyx/agollo/v2/storage"
	"github.com/zouyx/agollo/v2/utils"
)

const (
	long_poll_interval = 2 * time.Second //2s

	//notify timeout
	nofity_connect_timeout = 10 * time.Minute //10m
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

func init() {
	InitAllNotifications(nil)
}

func InitAllNotifications(callback func(namespace string)) {
	appConfig := env.GetPlainAppConfig()
	ns := env.SplitNamespaces(appConfig.NamespaceName, callback)
	allNotifications = &notificationsMap{
		notifications: ns,
	}
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

type NotifyConfigComponent struct {
}

func (this *NotifyConfigComponent) Start() {
	t2 := time.NewTimer(long_poll_interval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			NotifySyncConfigServices()
			t2.Reset(long_poll_interval)
		}
	}
}

func NotifySyncConfigServices() error {

	remoteConfigs, err := notifyRemoteConfig(nil, utils.Empty)

	if err != nil {
		return fmt.Errorf("notifySyncConfigServices: %s", err)
	}
	if len(remoteConfigs) == 0 {
		return fmt.Errorf("notifySyncConfigServices: empty remote config")
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	err = AutoSyncConfigServices(nil)

	//first sync fail then load config file
	appConfig := env.GetPlainAppConfig()
	if err != nil {
		env.SplitNamespaces(appConfig.NamespaceName, func(namespace string) {
			config, _ := env.LoadConfigFile(appConfig.BackupConfigPath, namespace)
			if config != nil {
				storage.UpdateApolloConfig(config, false)
			}
		})
	}
	//sync all config
	return nil
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

func notifyRemoteConfig(newAppConfig *config.AppConfig, namespace string) ([]*apolloNotify, error) {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyUrlSuffix(allNotifications.getNotifies(namespace), appConfig, newAppConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	notifies, err := http.RequestRecovery(appConfig, &env.ConnectConfig{
		Uri:     urlSuffix,
		Timeout: nofity_connect_timeout,
	}, &http.CallBack{
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

		allNotifications.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationId)
	}
}

func AutoSyncConfigServicesSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	apolloConfig, err := env.CreateApolloConfigWithJson(responseBody)

	if err != nil {
		Logger.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	appConfig := env.GetPlainAppConfig()

	storage.UpdateApolloConfig(apolloConfig, appConfig.GetIsBackupConfig())

	return nil, nil
}

func AutoSyncConfigServices(newAppConfig *config.AppConfig) error {
	return autoSyncNamespaceConfigServices(newAppConfig, allNotifications.notifications)
}

func autoSyncNamespaceConfigServices(newAppConfig *config.AppConfig, notifications map[string]int64) error {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	var err error
	for namespace := range notifications {
		urlSuffix := component.GetConfigURLSuffix(appConfig, namespace)

		_, err = http.RequestRecovery(appConfig, &env.ConnectConfig{
			Uri: urlSuffix,
		}, &http.CallBack{
			SuccessCallBack:   AutoSyncConfigServicesSuccessCallBack,
			NotModifyCallBack: touchApolloConfigCache,
		})
		if err != nil {
			return err
		}
	}
	return err
}

func getNotifyUrlSuffix(notifications string, config *config.AppConfig, newConfig *config.AppConfig) string {
	c := config
	if newConfig != nil {
		c = newConfig
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(c.AppId),
		url.QueryEscape(c.Cluster),
		url.QueryEscape(notifications))
}
