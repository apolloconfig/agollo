package http

import (
	"github.com/zouyx/agollo/v2/component/notify"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/component"
	"github.com/zouyx/agollo/v2/env"
)

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

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runNormalBackupConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := component.GetConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		Uri: urlSuffix,
	}, &CallBack{
		SuccessCallBack: notify.AutoSyncConfigServicesSuccessCallBack,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runLongTimeResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	startTime := time.Now().Unix()
	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := component.GetConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		Uri:     urlSuffix,
		Timeout: 11 * time.Second,
	}, &CallBack{
		SuccessCallBack: notify.AutoSyncConfigServicesSuccessCallBack,
	})

	endTime := time.Now().Unix()
	duration := endTime - startTime
	t.Log("start time:", startTime)
	t.Log("endTime:", endTime)
	t.Log("duration:", duration)
	Assert(t, int64(10), Equal(duration))
	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func mockIpList(t *testing.T) {
	server := runNormalServicesResponse()
	defer server.Close()
	time.Sleep(1 * time.Second)
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	err := env.SyncServerIpList(newAppConfig)

	Assert(t, err, NilVal())

	servers := env.GetServers()
	serverLen := 0
	servers.Range(func(k, v interface{}) bool {
		serverLen++
		return true
	})

	Assert(t, 2, Equal(serverLen))
}
