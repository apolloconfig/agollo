package http

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/env/config/json"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/utils"
)

var (
	jsonConfigFile = &json.ConfigFile{}
)

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

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIPList(t)
	server := runNormalBackupConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI: urlSuffix,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIPList(t)
	server := runLongTimeResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	startTime := time.Now().Unix()
	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI:     urlSuffix,
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

func mockIPList(t *testing.T) {
	time.Sleep(1 * time.Second)

	_, err := env.SyncServerIPListSuccessCallBack([]byte(servicesResponseStr))

	Assert(t, err, NilVal())

	serverLen := env.GetServersLen()

	Assert(t, 2, Equal(serverLen))
}

func getConfigURLSuffix(config *config.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(env.GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
}
