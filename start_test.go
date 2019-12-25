package agollo

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/component/notify"
	"github.com/zouyx/agollo/v2/env"

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

	value := getValue("key1")
	Assert(t, "value1", Equal(value))
}

func TestStartWithMultiNamespace(t *testing.T) {
	t.SkipNow()
	initDefaultConfig()
	notify.InitAllNotifications()
	app1 := "abc1"

	appConfig := env.GetPlainAppConfig()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap[defaultNamespace] = onlyNormalConfigResponse
	handlerMap[app1] = onlyNormalSecondConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalTwoResponse)
	appConfig.Ip = server.URL

	Start()

	time.Sleep(1 * time.Second)

	value := getValue("key1")
	Assert(t, "value1", Equal(value))

	config := GetConfig(app1)
	Assert(t, config, NotNilVal())
	Assert(t, config.getValue("key1-1"), Equal("value1-1"))
}

func TestErrorStart(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	time.Sleep(1 * time.Second)

	Start()

	value := getValue("key1")
	Assert(t, "value1", Equal(value))

	value2 := getValue("key2")
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