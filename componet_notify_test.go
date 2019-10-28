package agollo

import (
	"encoding/json"
	"fmt"
	. "github.com/tevid/gohamcrest"
	"testing"
	"time"
)

func TestSyncConfigServices(t *testing.T) {
	notifySyncConfigServices()
}

func TestGetRemoteConfig(t *testing.T) {
	server := runNormalResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	count := 1
	var remoteConfigs []*apolloNotify
	var err error
	for {
		count++
		remoteConfigs, err = notifyRemoteConfig(newAppConfig)

		//err keep nil
		Assert(t, err,NilVal())

		//if remote config is nil then break
		if remoteConfigs != nil && len(remoteConfigs) > 0 {
			break
		}
	}

	Assert(t, count > 1, Equal(true))
	Assert(t, err,NilVal())
	Assert(t, remoteConfigs,NotNilVal())
	Assert(t, 1, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	notify := remoteConfigs[0]

	Assert(t, "application", Equal(notify.NamespaceName))
	Assert(t, true, Equal(notify.NotificationId > 0))
}

func TestErrorGetRemoteConfig(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	appConfig.Ip=server.URL

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil)

	Assert(t, err,NotNilVal())
	Assert(t, remoteConfigs,NilVal())
	Assert(t, 0, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	Assert(t, "Over Max Retry Still Error!", Equal(err.Error()))
}

func initNotifications() {
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}
	allNotifications.notifications["application"]=-1
	allNotifications.notifications["abc1"]=-1
}

func TestUpdateAllNotifications(t *testing.T) {
	//clear
	initNotifications()

	notifyJson := `[
  {
    "namespaceName": "application",
    "notificationId": 101
  }
]`
	notifies := make([]*apolloNotify, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err,NilVal())
	Assert(t, true, Equal(len(notifies) > 0))

	updateAllNotifications(notifies)

	Assert(t, true, Equal(len(allNotifications.notifications) > 0))
	Assert(t, int64(101), Equal(allNotifications.notifications["application"]))
}

func TestUpdateAllNotificationsError(t *testing.T) {
	//clear
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}

	notifyJson := `ffffff`
	notifies := make([]*apolloNotify, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err,NotNilVal())
	Assert(t, true, Equal(len(notifies) == 0))

	updateAllNotifications(notifies)

	Assert(t, true, Equal(len(allNotifications.notifications) == 0))
}

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	Assert(t, notified,NilVal())
	Assert(t, err,NotNilVal())
}

func TestAutoSyncConfigServices(t *testing.T) {
	initNotifications()
	server := runNormalConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err := autoSyncConfigServices(newAppConfig)
	err = autoSyncConfigServices(newAppConfig)

	Assert(t, err,NilVal())

	config := GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	//Assert(t,"value1",config.Configurations["key1"])
	//Assert(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	server := runLongNotmodifiedConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	autoSyncConfigServicesSuccessCallBack([]byte(configResponseStr))

	config := GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	fmt.Println("sleeping 10s")

	time.Sleep(10 * time.Second)

	fmt.Println("checking agcache time left")
	defaultConfigCache := getDefaultConfigCache()

	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, string(value.([]byte)),NotNilVal())
		return true
	})

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(getValue("key1")))
	Assert(t, "value2", Equal(getValue("key2")))

	err := autoSyncConfigServices(newAppConfig)

	fmt.Println("checking agcache time left")
	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, string(value.([]byte)),NotNilVal())
		return true
	})

	fmt.Println(err)

	//sleep for async
	time.Sleep(1 * time.Second)
	checkBackupFile(t)
}

func checkBackupFile(t *testing.T) {
	newConfig, e := loadConfigFile(appConfig.getBackupConfigPath(),"application")
	t.Log(newConfig.Configurations)
	Assert(t,e,NilVal())
	Assert(t,newConfig.Configurations,NotNilVal())
	for k, v := range newConfig.Configurations {
		Assert(t, getValue(k), Equal(v))
	}
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go initFileConfig()
	server := runErrorConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	err := autoSyncConfigServices(nil)

	Assert(t, err,NotNilVal())

	config := GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	//still properties config
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "", Equal(config.ReleaseKey))
}
