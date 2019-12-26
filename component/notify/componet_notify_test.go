package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/storage"
)

func TestSyncConfigServices(t *testing.T) {
	//clear
	initNotifications()
	err := NotifySyncConfigServices()
	//err keep nil
	Assert(t, err, NilVal())
}

func TestGetRemoteConfig(t *testing.T) {
	//clear
	initNotifications()
	server := runNormalResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY)

	//err keep nil
	Assert(t, err, NilVal())

	Assert(t, err, NilVal())
	Assert(t, remoteConfigs, NotNilVal())
	Assert(t, 1, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	notify := remoteConfigs[0]

	Assert(t, "application", Equal(notify.NamespaceName))
	Assert(t, true, Equal(notify.NotificationId > 0))
}

func TestErrorGetRemoteConfig(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	appConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY)

	Assert(t, err, NotNilVal())
	Assert(t, remoteConfigs, NilVal())
	Assert(t, 0, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	Assert(t, "Over Max Retry Still Error!", Equal(err.Error()))
}

func initNotifications() {
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}
	allNotifications.notifications["application"] = -1
	allNotifications.notifications["abc1"] = -1
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

	Assert(t, err, NilVal())
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

	Assert(t, err, NotNilVal())
	Assert(t, true, Equal(len(notifies) == 0))

	updateAllNotifications(notifies)

	Assert(t, true, Equal(len(allNotifications.notifications) == 0))
}

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	Assert(t, notified, NilVal())
	Assert(t, err, NotNilVal())
}

func TestAutoSyncConfigServices(t *testing.T) {
	initNotifications()
	server := runNormalConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	env.GetPlainAppConfig().NextTryConnTime = 0

	err := AutoSyncConfigServices(newAppConfig)
	err = AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	//Assert(t,"value1",config.Configurations["key1"])
	//Assert(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesNoBackupFile(t *testing.T) {
	initNotifications()
	server := runNormalConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	appConfig := env.GetPlainAppConfig()
	appConfig.IsBackupConfig = false
	configFilePath := env.GetConfigFile(newAppConfig.GetBackupConfigPath(), "application")
	err := os.Remove(configFilePath)

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err = AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())
	checkNilBackupFile(t)
	appConfig.IsBackupConfig = true
}

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	server := runLongNotmodifiedConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	time.Sleep(1 * time.Second)
	appConfig := env.GetPlainAppConfig()
	appConfig.NextTryConnTime = 0

	AutoSyncConfigServicesSuccessCallBack([]byte(configResponseStr))

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	fmt.Println("sleeping 10s")

	time.Sleep(10 * time.Second)

	fmt.Println("checking agcache time left")
	defaultConfigCache := storage.GetDefaultConfigCache()

	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, string(value.([]byte)), NotNilVal())
		return true
	})

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(storage.GetStringValue("key1", "")))
	Assert(t, "value2", Equal(storage.GetStringValue("key2", "")))

	err := AutoSyncConfigServices(newAppConfig)

	fmt.Println("checking agcache time left")
	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, string(value.([]byte)), NotNilVal())
		return true
	})

	fmt.Println(err)

	//sleep for async
	time.Sleep(1 * time.Second)
	checkBackupFile(t)
}

func checkNilBackupFile(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	newConfig, e := env.LoadConfigFile(appConfig.GetBackupConfigPath(), "application")
	Assert(t, e, NotNilVal())
	Assert(t, newConfig, NilVal())
}

func checkBackupFile(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	newConfig, e := env.LoadConfigFile(appConfig.GetBackupConfigPath(), "application")
	t.Log(newConfig.Configurations)
	Assert(t, e, NilVal())
	Assert(t, newConfig.Configurations, NotNilVal())
	for k, v := range newConfig.Configurations {
		Assert(t, storage.GetStringValue(k, ""), Equal(v))
	}
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go env.InitConfig(nil)
	server := runErrorConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	err := AutoSyncConfigServices(nil)

	Assert(t, err, NotNilVal())

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	//still properties config
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "", Equal(config.ReleaseKey))
}

func getTestAppConfig() *env.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config, _ := env.CreateAppConfigWithJson(jsonStr)

	return config
}