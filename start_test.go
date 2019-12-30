package agollo

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/component/notify"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/storage"

	"net/http"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse)
	appConfig := env.GetPlainAppConfig()
	appConfig.Ip = server.URL

	Start()

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))
}

func TestStartWithMultiNamespace(t *testing.T) {
	t.SkipNow()
	storage.InitDefaultConfig()
	notify.InitAllNotifications()
	app1 := "abc1"

	appConfig := env.GetPlainAppConfig()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap[app1] = onlyNormalSecondConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalTwoResponse)
	appConfig.Ip = server.URL

	Start()

	time.Sleep(1 * time.Second)

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))

	config := storage.GetConfig(app1)
	Assert(t, config, NotNilVal())
	Assert(t, config.GetValue("key1-1"), Equal("value1-1"))
}

func TestErrorStart(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	Start()

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))

	value2 := GetValue("key2")
	Assert(t, "value2", Equal(value2))

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

func TestStructInit(t *testing.T) {

	readyConfig := &env.AppConfig{
		AppId:         "test1",
		Cluster:       "dev1",
		NamespaceName: "application1",
		Ip:            "localhost:8889",
	}

	InitCustomConfig(func() (*env.AppConfig, error) {
		return readyConfig, nil
	})

	time.Sleep(1 * time.Second)

	config := env.GetAppConfig(nil)
	Assert(t, config, NotNilVal())
	Assert(t, "test1", Equal(config.AppId))
	Assert(t, "dev1", Equal(config.Cluster))
	Assert(t, "application1", Equal(config.NamespaceName))
	Assert(t, "localhost:8889", Equal(config.Ip))

	apolloConfig := env.GetCurrentApolloConfig()[config.NamespaceName]
	Assert(t, "test1", Equal(apolloConfig.AppId))
	Assert(t, "dev1", Equal(apolloConfig.Cluster))
	Assert(t, "application1", Equal(apolloConfig.NamespaceName))

	//revert file config
	env.InitFileConfig()
}
