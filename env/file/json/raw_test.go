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
	"os"
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/qshuai/agollo/v4/extension"
)

func TestRawHandler_WriteConfigDirFile(t *testing.T) {
	extension.SetFileHandler(&rawFileHandler{})
	configPath := "raw-conf"
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": ["a", "b"]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.RemoveAll(configPath)

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
	os.RemoveAll(configPath)
}

func TestRawHandler_WriteConfigFile(t *testing.T) {
	extension.SetFileHandler(&rawFileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "test": ["a", "b"]
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.AppID, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestRawHandler_WriteConfigFileWithContent(t *testing.T) {
	extension.SetFileHandler(&rawFileHandler{})
	configPath := ""
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application.json",
  "configurations": {
    "content":"a: value1"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))
	Assert(t, err, NilVal())
	os.Remove(extension.GetFileHandler().GetConfigFile(configPath, config.AppID, config.NamespaceName))

	Assert(t, err, NilVal())
	e := extension.GetFileHandler().WriteConfigFile(config, configPath)
	Assert(t, e, NilVal())
}

func TestGetRawFileHandler(t *testing.T) {
	handler := GetRawFileHandler()
	Assert(t, handler, NotNilVal())

	fileHandler := GetRawFileHandler()
	Assert(t, handler, Equal(fileHandler))
}
