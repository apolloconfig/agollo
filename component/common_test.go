package component

import (
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
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

func TestSyncServerIpList(t *testing.T) {
	trySyncServerIpList(t)
}

func trySyncServerIpList(t *testing.T) {
	server := runMockServicesConfigServer()
	defer server.Close()

	newAppConfig := getTestAppConfig()
	newAppConfig.Ip = server.URL
	err := SyncServerIpList(newAppConfig)

	Assert(t, err, NilVal())

	servers := env.GetServers()
	serverLen := 0
	servers.Range(func(k, v interface{}) bool {
		serverLen++
		return true
	})

	Assert(t, 10, Equal(serverLen))

}

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

func TestSelectOnlyOneHost(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:" + appConfig.SelectHost())

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(appConfig.SelectHost()))
}

func TestSelectHost(t *testing.T) {
	//mock ip data
	trySyncServerIpList(t)

	servers := env.GetServers()
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:" + appConfig.SelectHost())

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(appConfig.SelectHost()))

	//check select next time
	appConfig.SetNextTryConnTime(5)
	Assert(t, host, NotEqual(appConfig.SelectHost()))
	time.Sleep(6 * time.Second)
	Assert(t, host, Equal(appConfig.SelectHost()))

	//check servers
	appConfig.SetNextTryConnTime(5)
	firstHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(firstHost))
	env.SetDownNode(firstHost)

	secondHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(secondHost))
	Assert(t, firstHost, NotEqual(secondHost))
	env.SetDownNode(secondHost)

	thirdHost := appConfig.SelectHost()
	Assert(t, host, NotEqual(thirdHost))
	Assert(t, firstHost, NotEqual(thirdHost))
	Assert(t, secondHost, NotEqual(thirdHost))

	servers.Range(func(k, v interface{}) bool {
		env.SetDownNode(k.(string))
		return true
	})

	Assert(t, "", Equal(appConfig.SelectHost()))

	//no servers
	//servers = make(map[string]*serverInfo, 0)
	deleteServers()
	Assert(t, "", Equal(appConfig.SelectHost()))
}

func deleteServers() {
	servers := env.GetServers()
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
}
