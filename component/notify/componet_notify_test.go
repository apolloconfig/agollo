package notify

import (
	"encoding/json"
	"fmt"
	"github.com/zouyx/agollo/v3/env/filehandler"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/zouyx/agollo/v3/env/config"
	jsonConfig "github.com/zouyx/agollo/v3/env/config/json"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env"
	_ "github.com/zouyx/agollo/v3/loadbalance/roundrobin"
)

var (
	jsonConfigFile = &jsonConfig.ConfigFile{}
	isAsync        = true
)

const responseStr = `[{"namespaceName":"application","notificationId":%d}]`

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configResponseStr)
}

func onlyNormalTwoConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configAbc1ResponseStr)
}

func onlyNormalResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}

func initMockNotifyAndConfigServer() {
	//clear
	initNotifications()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = onlyNormalTwoConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse)
	appConfig := env.GetPlainAppConfig()
	env.InitConfig(func() (*config.AppConfig, error) {
		appConfig.IP = server.URL
		appConfig.NextTryConnTime = 0
		return appConfig, nil
	})
}

func TestSyncConfigServices(t *testing.T) {
	initMockNotifyAndConfigServer()

	err := AsyncConfigs()
	//err keep nil
	Assert(t, err, NilVal())
}

func TestGetRemoteConfig(t *testing.T) {
	initMockNotifyAndConfigServer()

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY, isAsync)

	//err keep nil
	Assert(t, err, NilVal())

	Assert(t, err, NilVal())
	Assert(t, remoteConfigs, NotNilVal())
	Assert(t, 1, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	notify := remoteConfigs[0]

	Assert(t, "application", Equal(notify.NamespaceName))
	Assert(t, true, Equal(notify.NotificationID > 0))
}

func TestErrorGetRemoteConfig(t *testing.T) {
	//clear
	initNotifications()
	appConfig := env.GetPlainAppConfig()
	server := runErrorResponse()
	appConfig.IP = server.URL
	env.InitConfig(func() (*config.AppConfig, error) {
		appConfig.IP = server.URL
		appConfig.NextTryConnTime = 0
		return appConfig, nil
	})

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY, isAsync)

	Assert(t, err, NotNilVal())
	Assert(t, remoteConfigs, NilVal())
	Assert(t, 0, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	Assert(t, "over Max Retry Still Error", Equal(err.Error()))
}

func initNotifications() {
	allNotifications = &notificationsMap{
		notifications: sync.Map{},
	}
	allNotifications.setNotify("application", -1)
	allNotifications.setNotify("abc1", -1)
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

	Assert(t, true, Equal(allNotifications.GetNotifyLen() > 0))
	Assert(t, int64(101), Equal(allNotifications.getNotify("application")))
}

func TestUpdateAllNotificationsError(t *testing.T) {
	//clear
	allNotifications = &notificationsMap{
		notifications: sync.Map{},
	}

	notifyJson := `ffffff`
	notifies := make([]*apolloNotify, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err, NotNilVal())
	Assert(t, true, Equal(len(notifies) == 0))

	updateAllNotifications(notifies)

	Assert(t, true, Equal(allNotifications.GetNotifyLen() == 0))
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
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	env.GetPlainAppConfig().NextTryConnTime = 0

	err := AutoSyncConfigServices(newAppConfig)
	err = AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppID))
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
	newAppConfig.IP = server.URL
	appConfig := env.GetPlainAppConfig()
	appConfig.IsBackupConfig = false
	configFilePath := filehandler.GetFileHandler().GetConfigFile(newAppConfig.GetBackupConfigPath(), "application")
	err := os.Remove(configFilePath)

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err = AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())
	checkNilBackupFile(t)
	appConfig.IsBackupConfig = true
}

func checkNilBackupFile(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	newConfig, e := filehandler.GetFileHandler().LoadConfigFile(appConfig.GetBackupConfigPath(), "application")
	Assert(t, e, NotNilVal())
	Assert(t, newConfig, NilVal())
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go env.InitConfig(nil)
	server := runErrorConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	err := AutoSyncConfigServices(nil)

	Assert(t, err, NotNilVal())

	config := env.GetCurrentApolloConfig()[newAppConfig.NamespaceName]

	//still properties config
	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
}

func getTestAppConfig() *config.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	c, _ := env.Unmarshal([]byte(jsonStr))

	return c.(*config.AppConfig)
}
