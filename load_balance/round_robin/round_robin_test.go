package round_robin

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/load_balance"
	"testing"
	"time"
)

func TestSelectHost(t *testing.T) {
	balanace := load_balance.GetLoadBalanace()
	//mock ip data
	//trySyncServerIpList(t)

	servers := env.GetServers()
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:", balanace.Load(env.GetServers()).HomepageUrl)

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(balanace.Load(env.GetServers()).HomepageUrl))

	//check select next time
	appConfig.SetNextTryConnTime(5)
	Assert(t, host, NotEqual(balanace.Load(env.GetServers()).HomepageUrl))
	time.Sleep(6 * time.Second)
	Assert(t, host, Equal(balanace.Load(env.GetServers()).HomepageUrl))

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

	Assert(t, "", Equal(balanace.Load(env.GetServers()).HomepageUrl))

	//no servers
	//servers = make(map[string]*serverInfo, 0)
	deleteServers()
	Assert(t, "", Equal(balanace.Load(env.GetServers()).HomepageUrl))
}

func deleteServers() {
	servers := env.GetServers()
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
}
