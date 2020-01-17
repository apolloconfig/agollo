package env

import (
	"encoding/json"
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json_config"
	"github.com/zouyx/agollo/v2/utils"
	"os"
	"sync"

	"testing"
	"time"
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
	defaultNamespace = "application"
	jsonConfigFile   = &json_config.JSONConfigFile{}
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

func getTestAppConfig() *config.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	c, _ := Unmarshal([]byte(jsonStr))

	return c.(*config.AppConfig)
}

func TestLoadEnvConfig(t *testing.T) {
	envConfigFile := "env_test.properties"
	c, _ := jsonConfigFile.Load(APP_CONFIG_FILE_NAME, Unmarshal)
	config := c.(*config.AppConfig)
	config.Ip = "123"
	config.AppId = "1111"
	config.NamespaceName = "nsabbda"
	file, err := os.Create(envConfigFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(config)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = os.Setenv(ENV_CONFIG_FILE_PATH, envConfigFile)
	envConfig, envConfigErr := getLoadAppConfig(nil)
	t.Log(config)

	Assert(t, envConfigErr, NilVal())
	Assert(t, envConfig, NotNilVal())
	Assert(t, envConfig.AppId, Equal(config.AppId))
	Assert(t, envConfig.Cluster, Equal(config.Cluster))
	Assert(t, envConfig.NamespaceName, Equal(config.NamespaceName))
	Assert(t, envConfig.Ip, Equal(config.Ip))

	os.Remove(envConfigFile)
}

func TestGetPlainAppConfig(t *testing.T) {
	plainAppConfig := GetPlainAppConfig()
	Assert(t, plainAppConfig, NotNilVal())
}

func TestGetServersLen(t *testing.T) {
	servers.Store("a", "a")
	serversLen := GetServersLen()
	Assert(t, serversLen, Equal(1))
}

func TestSplitNamespaces(t *testing.T) {
	w := &sync.WaitGroup{}
	w.Add(3)
	namespaces := SplitNamespaces("a,b,c", func(namespace string) {
		w.Done()
	})

	Assert(t, len(namespaces), Equal(3))
	w.Wait()
}
func TestSyncServerIpListSuccessCallBack(t *testing.T) {
	SyncServerIpListSuccessCallBack([]byte(servicesConfigResponseStr))
	Assert(t, GetServersLen(), Equal(11))
}

func TestSetDownNode(t *testing.T) {
	t.SkipNow()
	SyncServerIpListSuccessCallBack([]byte(servicesConfigResponseStr))

	downNode := "10.15.128.102:8080"
	SetDownNode(downNode)

	value, ok := servers.Load("http://10.15.128.102:8080/")
	info := value.(*config.ServerInfo)
	Assert(t, ok, Equal(true))
	Assert(t, info.IsDown, Equal(true))
}
