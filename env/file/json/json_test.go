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

	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/extension"
	"github.com/qshuai/agollo/v4/utils"
)

func TestCreateDir(t *testing.T) {
	configPath := "conf"
	f := &FileHandler{}
	err := f.createDir(configPath)
	Assert(t, err, NilVal())
	err = f.createDir(configPath)
	Assert(t, err, NilVal())
	err = os.Mkdir(configPath, os.ModePerm)
	Assert(t, os.IsExist(err), Equal(true))

	err = f.createDir("")
	Assert(t, err, NilVal())

	os.RemoveAll(configPath)
}

func TestJSONFileHandler_WriteConfigDirFile(t *testing.T) {
	extension.SetFileHandler(&FileHandler{})
	configPath := "json-conf"
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": [1, 2]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.RemoveAll(configPath)
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.AppID, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
	os.RemoveAll(configPath)
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.AppID, config.NamespaceName))
}

func TestJSONFileHandler_WriteConfigFile(t *testing.T) {
	extension.SetFileHandler(&FileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": [1, 2]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.AppID, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestJSONFileHandler_LoadConfigFile(t *testing.T) {
	extension.SetFileHandler(&FileHandler{})
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": [1, 2]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NilVal())
	newConfig, e := extension.GetFileHandler().LoadConfigFile("", config.AppID, config.NamespaceName)

	t.Log(newConfig)
	Assert(t, e, NilVal())
	Assert(t, config.AppID, Equal(newConfig.AppID))
	Assert(t, config.ReleaseKey, Equal(newConfig.ReleaseKey))
	Assert(t, config.Cluster, Equal(newConfig.Cluster))
	Assert(t, config.NamespaceName, Equal(newConfig.NamespaceName))
}

func createApolloConfigWithJSON(b []byte) (*config.ApolloConfig, error) {
	apolloConfig := &config.ApolloConfig{}
	err := json.Unmarshal(b, apolloConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}
	return apolloConfig, nil
}
