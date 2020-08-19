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

package extension

import (
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/file"
)

type TestFileHandler struct {
}

//WriteConfigFile 写入配置文件
func (r *TestFileHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return nil
}

//GetConfigFile 获得配置文件路径
func (r *TestFileHandler) GetConfigFile(configDir string, namespace string) string {
	return ""
}

func (r *TestFileHandler) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	return nil, nil
}

func TestSetFileHandler(t *testing.T) {
	SetFileHandler(&TestFileHandler{})

	fileHandler := GetFileHandler()

	b := fileHandler.(file.FileHandler)
	Assert(t, b, NotNilVal())
}
