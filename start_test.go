package agollo

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/agcache/memory"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/component/notify"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	jsonFile "github.com/zouyx/agollo/v3/env/config/json"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/storage"
)

var (
	jsonConfigFile = &jsonFile.ConfigFile{}
	appConfigFile  = `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "backupConfigPath":""
}`
	appConfig = &config.AppConfig{
		AppID:         "test",
		Cluster:       "dev",
		NamespaceName: "application",
		IP:            "localhost:8888",
	}
)

func writeFile(content []byte, configPath string) {
	file, e := os.Create(configPath)
	if e != nil {
		log.Errorf("writeConfigFile fail,error:", e)
	}
	defer file.Close()
	file.Write(content)
}

func TestStart(t *testing.T) {
	c := appConfig
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse, c)
	c.IP = server.URL

	b, _ := json.Marshal(c)
	writeFile(b, "app.properties")

	Start()

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))
	handler := extension.GetFileHandler()
	Assert(t, handler, NotNilVal())
}

func TestStartWithMultiNamespace(t *testing.T) {
	notify.InitAllNotifications(nil)
	c := appConfig
	app1 := "abc1"

	appConfig := env.GetPlainAppConfig()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap[app1] = onlyNormalSecondConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalTwoResponse, appConfig)

	c.NamespaceName = "application,abc1"
	c.IP = server.URL
	b, _ := json.Marshal(c)
	writeFile(b, "app.properties")

	Start()

	time.Sleep(1 * time.Second)

	value := GetValue("key1")
	Assert(t, "value1", Equal(value))

	time.Sleep(1 * time.Second)
	config := storage.GetConfig(app1)
	Assert(t, config, NotNilVal())
	Assert(t, config.GetValue("key1-1"), Equal("value1-1"))

	rollbackFile()
}

func rollbackFile() {
	writeFile([]byte(appConfigFile), "app.properties")
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
	initAppConfigFunc = nil
	f := func() (*config.AppConfig, error) {
		return appConfig, nil
	}
	InitCustomConfig(f)
	Assert(t, initAppConfigFunc, NotNilVal())
}

func TestSetLogger(t *testing.T) {
	logger := &log.DefaultLogger{}
	SetLogger(logger)
	Assert(t, log.Logger, Equal(logger))
}

func TestUseEventDispatch(t *testing.T) {
	UseEventDispatch()
	l := storage.GetChangeListeners()
	Assert(t, l.Len(), Equal(1))
}

func TestSetCache(t *testing.T) {
	defaultCacheFactory := &memory.DefaultCacheFactory{}
	SetCache(defaultCacheFactory)
	Assert(t, extension.GetCacheFactory(), Equal(defaultCacheFactory))
}

type TestLoadBalance struct {
}

//Load 负载均衡
func (r *TestLoadBalance) Load(servers *sync.Map) *config.ServerInfo {
	return nil
}

func TestSetLoadBalance(t *testing.T) {
	balance := extension.GetLoadBalance()
	Assert(t, balance, NotNilVal())

	t2 := &TestLoadBalance{}
	SetLoadBalance(t2)
	Assert(t, t2, Equal(extension.GetLoadBalance()))
}

//testFileHandler 默认备份文件读写
type testFileHandler struct {
}

// WriteConfigFile write config to file
func (fileHandler *testFileHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return nil
}

// GetConfigFile get real config file
func (fileHandler *testFileHandler) GetConfigFile(configDir string, namespace string) string {
	return ""
}

// LoadConfigFile load config from file
func (fileHandler *testFileHandler) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	return nil, nil
}

func TestSetBackupFileHandler(t *testing.T) {
	fileHandler := extension.GetFileHandler()
	Assert(t, fileHandler, NotNilVal())

	t2 := &testFileHandler{}
	SetBackupFileHandler(t2)
	Assert(t, t2, Equal(extension.GetFileHandler()))
}

type TestAuth struct{}

func (a *TestAuth) HTTPHeaders(url string, appID string, secret string) map[string][]string {
	return nil
}

func TestSetSignature(t *testing.T) {
	Assert(t, extension.GetHTTPAuth(), NotNilVal())

	t2 := &TestAuth{}
	SetSignature(t2)

	Assert(t, t2, Equal(extension.GetHTTPAuth()))
}
