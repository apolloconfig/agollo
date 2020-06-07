package component

import (
	"testing"

	. "github.com/tevid/gohamcrest"
	_ "github.com/zouyx/agollo/v3/cluster/roundrobin"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/env/config/json"
	"github.com/zouyx/agollo/v3/extension"
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

func TestSelectOnlyOneHost(t *testing.T) {
	trySyncServerIPList()
	appConfig := env.GetPlainAppConfig()
	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	load := extension.GetLoadBalance().Load(env.GetServers())
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
