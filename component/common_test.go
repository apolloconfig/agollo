package component

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json_config"
	"github.com/zouyx/agollo/v2/load_balance"
	"testing"
)

var (
	jsonConfigFile = &json_config.JSONConfigFile{}
)

func TestCreateApolloConfigWithJson(t *testing.T) {
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := env.CreateApolloConfigWithJson([]byte(jsonStr))

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())

	Assert(t, "100004458", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(config.Configurations["key1"]))
	Assert(t, "value2", Equal(config.Configurations["key2"]))

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config, err := env.CreateApolloConfigWithJson([]byte(jsonStr))

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}

func TestSelectOnlyOneHost(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:" + load_balance.GetLoadBalanace().Load(env.GetServers()).HomepageUrl)

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(load_balance.GetLoadBalanace().Load(env.GetServers()).HomepageUrl))
}

func TestGetConfigURLSuffix(t *testing.T) {
	appConfig := &config.AppConfig{}
	uri := GetConfigURLSuffix(appConfig, "kk")
	Assert(t, "", NotEqual(uri))
}
