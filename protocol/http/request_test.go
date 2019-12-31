package http

import (
	"fmt"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json_config"
	"net/url"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/utils"
)

var(
	jsonConfigFile = &json_config.JSONConfigFile{}
)

func getTestAppConfig() *config.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config, _ := jsonConfigFile.Unmarshal(jsonStr)

	return config
}

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIpList(t)
	server := runNormalBackupConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL

	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		Uri: urlSuffix,
	}, &CallBack{
		SuccessCallBack: nil,
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
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		Uri:     urlSuffix,
		Timeout: 11 * time.Second,
	}, &CallBack{
		SuccessCallBack: nil,
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
	time.Sleep(1 * time.Second)

	_, err := env.SyncServerIpListSuccessCallBack([]byte(servicesResponseStr))

	Assert(t, err, NilVal())

	serverLen := env.GetServersLen()

	Assert(t, 2, Equal(serverLen))
}

func getConfigURLSuffix(config *config.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(env.GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
}
