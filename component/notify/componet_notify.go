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
	longPollInterval = 2 * time.Second //2s

	//notify timeout
	nofityConnectTimeout = 10 * time.Minute //10m

	//同步链接时间
	syncNofityConnectTimeout = 3 * time.Second //3s
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

//InitAllNotifications 初始化notificationsMap
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

	remoteConfigs, err := notifyRemoteConfig(nil, utils.Empty, isAsync)

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
		config, _ := env.LoadConfigFile(appConfig.BackupConfigPath, namespace)
		if config != nil {
			storage.UpdateApolloConfig(config, false)
		}
	})
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

func notifyRemoteConfig(newAppConfig *config.AppConfig, namespace string, isAsync bool) ([]*apolloNotify, error) {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix := getNotifyURLSuffix(allNotifications.getNotifies(namespace), appConfig, newAppConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	connectConfig := &env.ConnectConfig{
		URI: urlSuffix,
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

		allNotifications.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationId)
	}
}

//AutoSyncConfigServicesSuccessCallBack 同步配置回调
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

//AutoSyncConfigServices 自动同步配置
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
			URI: urlSuffix,
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

func getNotifyURLSuffix(notifications string, config *config.AppConfig, newConfig *config.AppConfig) string {
	c := config
	if newConfig != nil {
		c = newConfig
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(c.AppId),
		url.QueryEscape(c.Cluster),
		url.QueryEscape(notifications))
}
