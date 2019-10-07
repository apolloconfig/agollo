package agollo

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/test"
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
		test.Nil(t, err)

		//if remote config is nil then break
		if remoteConfigs != nil && len(remoteConfigs) > 0 {
			break
		}
	}

	test.Equal(t, count > 1, true)
	test.Nil(t, err)
	test.NotNil(t, remoteConfigs)
	test.Equal(t, 1, len(remoteConfigs))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	notify := remoteConfigs[0]

	test.Equal(t, "application", notify.NamespaceName)
	test.Equal(t, true, notify.NotificationId > 0)
}

func TestErrorGetRemoteConfig(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil)

	test.NotNil(t, err)
	test.Nil(t, remoteConfigs)
	test.Equal(t, 0, len(remoteConfigs))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	test.Equal(t, "Over Max Retry Still Error!", err.Error())
}

func TestUpdateAllNotifications(t *testing.T) {
	//clear
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}
	notifyJson := `[
  {
    "namespaceName": "application",
    "notificationId": 101
  }
]`
	notifies := make([]*apolloNotify, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	test.Nil(t, err)
	test.Equal(t, true, len(notifies) > 0)

	updateAllNotifications(notifies)

	test.Equal(t, true, len(allNotifications.notifications) > 0)
	test.Equal(t, int64(101), allNotifications.notifications["application"])
}

func TestUpdateAllNotificationsError(t *testing.T) {
	//clear
	allNotifications = &notificationsMap{
		notifications: make(map[string]int64, 1),
	}

	notifyJson := `ffffff`
	notifies := make([]*apolloNotify, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	test.NotNil(t, err)
	test.Equal(t, true, len(notifies) == 0)

	updateAllNotifications(notifies)

	test.Equal(t, true, len(allNotifications.notifications) == 0)
}

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	test.Nil(t, notified)
	test.NotNil(t, err)
}

func TestAutoSyncConfigServices(t *testing.T) {
	server := runNormalConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err := autoSyncConfigServices(newAppConfig)
	err = autoSyncConfigServices(newAppConfig)

	test.Nil(t, err)

	config := GetCurrentApolloConfig()

	test.Equal(t, "100004458", config.AppId)
	test.Equal(t, "default", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "20170430092936-dee2d58e74515ff3", config.ReleaseKey)
	//test.Equal(t,"value1",config.Configurations["key1"])
	//test.Equal(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	server := runLongNotmodifiedConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	autoSyncConfigServicesSuccessCallBack([]byte(configResponseStr))

	config := GetCurrentApolloConfig()

	fmt.Println("sleeping 10s")

	time.Sleep(10 * time.Second)

	fmt.Println("checking agcache time left")
	defaultConfigCache := getDefaultConfigCache()
	it := defaultConfigCache.NewIterator()
	for i := int64(0); i < defaultConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry == nil {
			break
		}
		timeLeft, err := defaultConfigCache.TTL([]byte(entry.Key))
		test.Nil(t, err)
		fmt.Printf("key:%s,time:%v \n", string(entry.Key), timeLeft)
		test.Equal(t, timeLeft >= 110, true)
	}

	test.Equal(t, "100004458", config.AppId)
	test.Equal(t, "default", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "20170430092936-dee2d58e74515ff3", config.ReleaseKey)
	test.Equal(t, "value1", getValue("key1"))
	test.Equal(t, "value2", getValue("key2"))

	err := autoSyncConfigServices(newAppConfig)

	fmt.Println("checking agcache time left")
	it1 := defaultConfigCache.NewIterator()
	for i := int64(0); i < defaultConfigCache.EntryCount(); i++ {
		entry := it1.Next()
		if entry == nil {
			break
		}
		timeLeft, err := defaultConfigCache.TTL([]byte(entry.Key))
		test.Nil(t, err)
		fmt.Printf("key:%s,time:%v \n", string(entry.Key), timeLeft)
		test.Equal(t, timeLeft >= 120, true)
	}

	fmt.Println(err)

	//sleep for async
	time.Sleep(1 * time.Second)
	checkBackupFile(t)
}

func checkBackupFile(t *testing.T) {
	newConfig, e := loadConfigFile(appConfig.getBackupConfigPath())
	t.Log(newConfig.Configurations)
	isNil(e)
	isNotNil(newConfig.Configurations)
	for k, v := range newConfig.Configurations {
		test.Equal(t, getValue(k), v)
	}
}

//test if not modify
func TestAutoSyncConfigServicesNotModify(t *testing.T) {
	server := runNotModifyConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	apolloConfig, err := createApolloConfigWithJson([]byte(configResponseStr))
	updateApolloConfig(apolloConfig, true)

	time.Sleep(10 * time.Second)
	checkCacheLeft(t, configCacheExpireTime-10)

	appConfig.NextTryConnTime = 0

	err = autoSyncConfigServices(newAppConfig)

	test.Nil(t, err)

	config := GetCurrentApolloConfig()

	test.Equal(t, "100004458", config.AppId)
	test.Equal(t, "default", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "20170430092936-dee2d58e74515ff3", config.ReleaseKey)

	checkCacheLeft(t, configCacheExpireTime)

	//test.Equal(t,"value1",config.Configurations["key1"])
	//test.Equal(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go initFileConfig()
	server := runErrorConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	err := autoSyncConfigServices(nil)

	test.NotNil(t, err)

	config := GetCurrentApolloConfig()

	//still properties config
	test.Equal(t, "test", config.AppId)
	test.Equal(t, "dev", config.Cluster)
	test.Equal(t, "application", config.NamespaceName)
	test.Equal(t, "", config.ReleaseKey)
}
