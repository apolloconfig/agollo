package json_config

import (
	"testing"

	. "github.com/tevid/gohamcrest"
)

var(
	APP_CONFIG_FILE_NAME = "app.properties"
	jsonConfigFile = &JSONConfigFile{}
)

func TestLoadJsonConfig(t *testing.T) {
	config, err := jsonConfigFile.LoadJsonConfig(APP_CONFIG_FILE_NAME)
	t.Log(config)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application,abc1", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.Ip))

}

func TestLoadJsonConfigWrongFile(t *testing.T) {
	config, err := jsonConfigFile.LoadJsonConfig("")
	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())

	Assert(t, err.Error(), StartWith("Fail to read config file"))
}

func TestLoadJsonConfigWrongType(t *testing.T) {
	config, err := jsonConfigFile.LoadJsonConfig("json_config.go")
	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())

	Assert(t, err.Error(), StartWith("Load Json Config fail"))
}

func TestCreateAppConfigWithJson(t *testing.T) {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": ""
	}`
	config, err := jsonConfigFile.Unmarshal(jsonStr)
	t.Log(config)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.Ip))
}

//func TestCreateAppConfigWithJsonWrongEnv(t *testing.T) {
//	jsonStr:=`{
//    "appId": "test",
//    "cluster": "joe",
//    "namespaceName": "application",
//    "ip": "localhost:8888",
//    "releaseKey": ""
//	}`
//	config,err:=createAppConfigWithJson(jsonStr)
//	t.Log(config)
//	t.Log(err)
//
//	Assert(t,err)
//	Assert(t,config)
//	test.StartWith(t,"Env is wrong ,current env:joe",err.Error())
//}

func TestCreateAppConfigWithJsonError(t *testing.T) {
	jsonStr := `package agollo

import (
	"os"
	"strconv"
	"time"
	"fmt"
	"net/url"
)`
	config, err := jsonConfigFile.Unmarshal(jsonStr)
	t.Log(err)

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}

func TestCreateAppConfigWithJsonDefault(t *testing.T) {
	jsonStr := `{
    "appId": "testDefault",
    "ip": "localhost:9999"
	}`
	config, err := jsonConfigFile.Unmarshal(jsonStr)
	t.Log(err)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "testDefault", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "localhost:9999", Equal(config.Ip))
}