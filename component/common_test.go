package component

import (
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json_config"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env"
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

func TestSelectOnlyOneHost(t *testing.T) {
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:" + appConfig.SelectHost(env.GetServers()))

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(appConfig.SelectHost(env.GetServers())))
}

func TestSelectHost(t *testing.T) {
	//mock ip data
	trySyncServerIpList(t)

	servers := env.GetServers()
	appConfig := env.GetPlainAppConfig()
	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:" + appConfig.SelectHost(env.GetServers()))

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, Equal(appConfig.SelectHost(env.GetServers())))

	//check select next time
	appConfig.SetNextTryConnTime(5)
	Assert(t, host, NotEqual(appConfig.SelectHost(env.GetServers())))
	time.Sleep(6 * time.Second)
	Assert(t, host, Equal(appConfig.SelectHost(env.GetServers())))

	//check servers
	appConfig.SetNextTryConnTime(5)
	firstHost := appConfig.SelectHost(env.GetServers())
	Assert(t, host, NotEqual(firstHost))
	env.SetDownNode(firstHost)

	secondHost := appConfig.SelectHost(env.GetServers())
	Assert(t, host, NotEqual(secondHost))
	Assert(t, firstHost, NotEqual(secondHost))
	env.SetDownNode(secondHost)

	thirdHost := appConfig.SelectHost(env.GetServers())
	Assert(t, host, NotEqual(thirdHost))
	Assert(t, firstHost, NotEqual(thirdHost))
	Assert(t, secondHost, NotEqual(thirdHost))

	servers.Range(func(k, v interface{}) bool {
		env.SetDownNode(k.(string))
		return true
	})

	Assert(t, "", Equal(appConfig.SelectHost(env.GetServers())))

	//no servers
	//servers = make(map[string]*serverInfo, 0)
	deleteServers()
	Assert(t, "", Equal(appConfig.SelectHost(env.GetServers())))
}

func deleteServers() {
	servers := env.GetServers()
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
}

func TestGetConfigURLSuffix(t *testing.T) {
	appConfig := &config.AppConfig{}
	uri := GetConfigURLSuffix(appConfig, "kk")
	Assert(t, "", NotEqual(uri))
}
