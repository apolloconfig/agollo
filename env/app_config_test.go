package env

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/utils"

	"testing"
	"time"
)

var (
	defaultNamespace = "application"
)

func TestInit(t *testing.T) {
	config := GetAppConfig(nil)
	time.Sleep(1 * time.Second)

	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application,abc1", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.Ip))

	//TODO: 需要确认是否放在这里
	//defaultApolloConfig := GetCurrentApolloConfig()[defaultNamespace]
	//Assert(t, defaultApolloConfig, NotNilVal())
	//Assert(t, "test", Equal(defaultApolloConfig.AppId))
	//Assert(t, "dev", Equal(defaultApolloConfig.Cluster))
	//Assert(t, "application", Equal(defaultApolloConfig.NamespaceName))
}

func TestGetServicesConfigUrl(t *testing.T) {
	appConfig := getTestAppConfig()
	url := GetServicesConfigUrl(appConfig)
	ip := utils.GetInternal()
	Assert(t, "http://localhost:8888/services/config?appId=test&ip="+ip, Equal(url))
}

func getTestAppConfig() *AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	config, _ := CreateAppConfigWithJson(jsonStr)

	return config
}
