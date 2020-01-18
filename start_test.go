package agollo

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/agcache"
	"github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/component/notify"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/env/config/json"
	"github.com/zouyx/agollo/v2/storage"

	"net/http"
	"testing"
	"time"
)

var (
	jsonConfigFile = &json.ConfigFile{}
)

func TestStart(t *testing.T) {
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse)
	appConfig := env.GetPlainAppConfig()
	appConfig.IP = server.URL

	Start()

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))
}

func TestStartWithMultiNamespace(t *testing.T) {
	env.GetPlainAppConfig().NamespaceName = "application,abc1"
	notify.InitAllNotifications(nil)
	app1 := "abc1"

	appConfig := env.GetPlainAppConfig()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap[app1] = onlyNormalSecondConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalTwoResponse)
	appConfig.IP = server.URL

	Start()

	time.Sleep(1 * time.Second)

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))

	time.Sleep(1 * time.Second)
	config := storage.GetConfig(app1)
	Assert(t, config, NotNilVal())
	Assert(t, config.GetValue("key1-1"), Equal("value1-1"))
}

func TestErrorStart(t *testing.T) {
	t.SkipNow()
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL
	notify.InitAllNotifications(nil)

	time.Sleep(1 * time.Second)

	Start()

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))

	value2 := GetValue("key2")
	Assert(t, "value2", Equal(value2))

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

func TestStructInit(t *testing.T) {
	t.SkipNow()
	readyConfig := &config.AppConfig{
		AppID:         "test1",
		Cluster:       "dev1",
		NamespaceName: "application1",
		IP:            "localhost:8889",
	}

	InitCustomConfig(func() (*config.AppConfig, error) {
		return readyConfig, nil
	})
	notify.InitAllNotifications(nil)

	time.Sleep(1 * time.Second)

	config := env.GetAppConfig(nil)
	Assert(t, config, NotNilVal())
	Assert(t, "test1", Equal(config.AppID))
	Assert(t, "dev1", Equal(config.Cluster))
	Assert(t, "application1", Equal(config.NamespaceName))
	Assert(t, "localhost:8889", Equal(config.IP))

	apolloConfig := env.GetCurrentApolloConfig()[config.NamespaceName]
	Assert(t, "test1", Equal(apolloConfig.AppID))
	Assert(t, "dev1", Equal(apolloConfig.Cluster))
	Assert(t, "application1", Equal(apolloConfig.NamespaceName))

	//revert file config
	env.InitFileConfig()
}

func TestInitCustomConfig(t *testing.T) {
	appConfig := &config.AppConfig{}
	InitCustomConfig(func() (*config.AppConfig, error) {
		return appConfig, nil
	})
	Assert(t, env.GetPlainAppConfig(), Equal(appConfig))
}

func TestSetLogger(t *testing.T) {
	logger := &log.DefaultLogger{}
	SetLogger(logger)
	Assert(t, log.Logger, Equal(logger))
}

func TestSetCache(t *testing.T) {
	defaultCacheFactory := &agcache.DefaultCacheFactory{}
	SetCache(defaultCacheFactory)
	Assert(t, agcache.GetCacheFactory(), Equal(defaultCacheFactory))
}
