/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package json

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/utils"
)

var (
	appConfigFile  = "app.properties"
	jsonConfigFile = &ConfigFile{}
)

func TestLoadJsonConfig(t *testing.T) {
	c, err := jsonConfigFile.Load(appConfigFile, Unmarshal)
	config := c.(*config.AppConfig)
	t.Log(config)

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())
	Assert(t, "test", Equal(config.AppID))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application,abc1", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.IP))

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
	Assert(t, "test", Equal(config.AppID))
	Assert(t, "dev", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "localhost:8888", Equal(config.IP))
}

// func TestCreateAppConfigWithJsonWrongEnv(t *testing.T) {
// 	jsonStr:=`{
//    "appId": "test",
//    "cluster": "joe",
//    "namespaceName": "application",
//    "ip": "localhost:8888",
//    "releaseKey": ""
// 	}`
// 	config,err:=createAppConfigWithJson(jsonStr)
// 	t.Log(config)
// 	t.Log(err)
//
// 	Assert(t,err)
// 	Assert(t,config)
// 	test.StartWith(t,"Env is wrong ,current env:joe",err.Error())
// }

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
	Assert(t, "testDefault", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "localhost:9999", Equal(config.IP))
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

func TestJSONConfigFile_Write(t *testing.T) {
	fileName := "test.json"
	jsonConfigFile.Write(`{"appId":"100004458","cluster":"default","namespaceName":"application","releaseKey":"20170430092936-dee2d58e74515ff3","configurations":{"key1":"value1","key2":"value2"}}`, fileName)
	file, e := os.Open(fileName)
	Assert(t, e, NilVal())
	Assert(t, file, NotNilVal())
	file.Close()
	os.Remove(fileName)
}

func TestJSONConfigFile_Write_error(t *testing.T) {
	fileName := "/a/a/a/a//s.k"
	e := jsonConfigFile.Write(`{"appId":"100004458","cluster":"default","namespaceName":"application","releaseKey":"20170430092936-dee2d58e74515ff3","configurations":{"key1":"value1","key2":"value2"}}`, fileName)
	file, _ := os.Open(fileName)
	Assert(t, e, NotNilVal())
	Assert(t, file, NilVal())

	e = jsonConfigFile.Write(``, fileName)
	file, _ = os.Open(fileName)
	Assert(t, e, NotNilVal())
	Assert(t, file, NilVal())
}
