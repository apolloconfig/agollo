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
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/apolloconfig/agollo/v4/env/config"

	. "github.com/tevid/gohamcrest"
)

type TestFileHandler struct {
}

// WriteConfigFile 写入配置文件
func (r *TestFileHandler) WriteConfigFile(config *config.ApolloConfig, configPath string) error {
	return nil
}

// GetConfigFile 获得配置文件路径
func (r *TestFileHandler) GetConfigFile(configDir string, appID string, namespace string) string {
	return ""
}

func (r *TestFileHandler) LoadConfigFile(configDir string, appID string, namespace string, cluster string) (*config.ApolloConfig, error) {
	return nil, nil
}

func TestSetFileHandler(t *testing.T) {
	AddFileHandler(&TestFileHandler{})

	fileHandler := GetFileHandlers()[0].(*TestFileHandler)

	Assert(t, fileHandler, NotNilVal())
}

func TestAddAndGetFileHandlers(t *testing.T) {
	// 清理handlers切片，确保测试环境干净
	handlers = []handlerWithPriority{}

	handler1 := &TestFileHandler{}
	handler2 := &TestFileHandler{}
	handler3 := &TestFileHandler{}

	// 添加三个处理器，具有不同的优先级
	AddFileHandler(handler1, 1)
	AddFileHandler(handler2, 3)
	AddFileHandler(handler3, 2)

	// 获取并验证处理器的顺序
	sortedHandlers := GetFileHandlers()
	assert.Equal(t, 3, len(sortedHandlers), "应该有三个处理器")
	assert.IsType(t, handler2, sortedHandlers[0], "优先级最高的处理器应该在第一位")
	assert.IsType(t, handler3, sortedHandlers[1], "第二高优先级的处理器应该在第二位")
	assert.IsType(t, handler1, sortedHandlers[2], "优先级最低的处理器应该在最后一位")
}
