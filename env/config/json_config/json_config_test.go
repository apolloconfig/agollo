package json_config

import (
	"encoding/json"
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/utils"
	"testing"

	. "github.com/tevid/gohamcrest"
)

var (
	APP_CONFIG_FILE_NAME = "app.properties"
	jsonConfigFile       = &JSONConfigFile{}
)

func TestLoadJsonConfig(t *testing.T) {
	c, err := jsonConfigFile.Load(APP_CONFIG_FILE_NAME, Unmarshal)
	config := c.(*config.AppConfig)
	t.Log(config)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppId))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application,abc1", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.Ip))

}

func TestLoadJsonConfigWrongFile(t *testing.T) {
	config, err := jsonConfigFile.Load("", Unmarshal)
	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())

	Assert(t, err.Error(), StartWith("Fail to read config file"))
}

func TestLoadJsonConfigWrongType(t *testing.T) {
	config, err := jsonConfigFile.Load("json_config.go", Unmarshal)
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
	c, err := Unmarshal([]byte(jsonStr))
	config := c.(*config.AppConfig)
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
	config, err := Unmarshal([]byte(jsonStr))
	t.Log(err)

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}

func TestCreateAppConfigWithJsonDefault(t *testing.T) {
	jsonStr := `{
    "appId": "testDefault",
    "ip": "localhost:9999"
	}`
	c, err := Unmarshal([]byte(jsonStr))
	config := c.(*config.AppConfig)
	t.Log(err)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "testDefault", Equal(config.AppId))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "localhost:9999", Equal(config.Ip))
}

func Unmarshal(b []byte) (interface{}, error) {
	appConfig := &config.AppConfig{
		Cluster:        "default",
		NamespaceName:  "application",
		IsBackupConfig: true,
	}
	err := json.Unmarshal(b, appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	return appConfig, nil
}
