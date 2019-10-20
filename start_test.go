package agollo

import (
	. "github.com/tevid/gohamcrest"
	"net/http"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"]=onlyNormalConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse)
	appConfig.Ip = server.URL

	Start()

	value := getValue("key1")
	Assert(t, "value1", Equal(value))
}

func TestStartWithMultiNamespace(t *testing.T) {
	t.SkipNow()
	initDefaultConfig()
	initNotifications()
	app1 := "abc1"

	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap[defaultNamespace]=onlyNormalConfigResponse
	handlerMap[app1]=onlyNormalSecondConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalTwoResponse)
	appConfig.Ip = server.URL

	Start()

	time.Sleep(1* time.Second)

	value := getValue("key1")
	Assert(t, "value1", Equal(value))

	config := GetConfig(app1)
	Assert(t,config,NotNilVal())
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
