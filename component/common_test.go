package component

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/env/config/json"
	"github.com/zouyx/agollo/v3/loadbalance"
	_ "github.com/zouyx/agollo/v3/loadbalance/roundrobin"
	"testing"
)

const servicesConfigResponseStr = `[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.102:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.125:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.11:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.11:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.193:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.193:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.101:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.101:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.192:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.192:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.124:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.124:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.103:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.103:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "localhost:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.12:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.194:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.194:8080/"
}
]`

var (
	jsonConfigFile = &json.ConfigFile{}
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

	config, err := env.CreateApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())

	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(config.Configurations["key1"]))
	Assert(t, "value2", Equal(config.Configurations["key2"]))

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config, err := env.CreateApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}

func TestSelectOnlyOneHost(t *testing.T) {
	trySyncServerIPList()
	appConfig := env.GetPlainAppConfig()
	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	load := loadbalance.GetLoadBalance().Load(env.GetServers())
	Assert(t, load, NotNilVal())
	Assert(t, host, NotEqual(load.HomepageURL))
}

func TestGetConfigURLSuffix(t *testing.T) {
	appConfig := &config.AppConfig{}
	uri := GetConfigURLSuffix(appConfig, "kk")
	Assert(t, "", NotEqual(uri))

	uri = GetConfigURLSuffix(nil, "kk")
	Assert(t, "", Equal(uri))
}

type testComponent struct {
}

//Start 启动同步服务器列表
func (s *testComponent) Start() {
}

func TestStartRefreshConfig(t *testing.T) {
	StartRefreshConfig(&testComponent{})
}

func TestName(t *testing.T) {

}

func trySyncServerIPList() {
	env.SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr))
}
