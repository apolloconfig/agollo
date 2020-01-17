package roundrobin

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/loadbalance"
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

func TestSelectHost(t *testing.T) {
	balanace := loadbalance.GetLoadBalance()
	//mock ip data
	trySyncServerIPList()

	servers := env.GetServers()
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:", balanace.Load(env.GetServers()).HomepageUrl)

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, NotEqual(balanace.Load(env.GetServers()).HomepageUrl))

	//check select next time
	appConfig.SetNextTryConnTime(5)
	Assert(t, host, NotEqual(balanace.Load(env.GetServers()).HomepageUrl))

	//check servers
	appConfig.SetNextTryConnTime(5)
	firstHost := balanace.Load(env.GetServers())
	Assert(t, host, NotEqual(firstHost.HomepageUrl))
	env.SetDownNode(firstHost.HomepageUrl)

	secondHost := balanace.Load(env.GetServers()).HomepageUrl
	Assert(t, host, NotEqual(secondHost))
	Assert(t, firstHost, NotEqual(secondHost))
	env.SetDownNode(secondHost)

	thirdHost := balanace.Load(env.GetServers()).HomepageUrl
	Assert(t, host, NotEqual(thirdHost))
	Assert(t, firstHost, NotEqual(thirdHost))
	Assert(t, secondHost, NotEqual(thirdHost))

	servers.Range(func(k, v interface{}) bool {
		env.SetDownNode(k.(string))
		return true
	})

	Assert(t, balanace.Load(env.GetServers()), NilVal())

	//no servers
	//servers = make(map[string]*serverInfo, 0)
	deleteServers()
	Assert(t, balanace.Load(env.GetServers()), NilVal())
}

func deleteServers() {
	servers := env.GetServers()
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
}

func trySyncServerIPList() {
	env.SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr))
}
