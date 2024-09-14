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
	"github.com/apolloconfig/agollo/v4/env/config"
	. "github.com/tevid/gohamcrest"
	"testing"
)

type TestConfigMapHandler struct {
}

func (t *TestConfigMapHandler) LoadConfigMap(configMapNamespace string) (*config.ApolloConfig, error) {
	return nil, nil
}

func (t *TestConfigMapHandler) WriteConfigMap(config *config.ApolloConfig, configMapNamespace string) error {
	return nil
}

// TestSetConfigMapHandler 测试 SetConfigMapHandler 函数
func TestSetConfigMapHandler(t *testing.T) {
	SetConfigMapHandler(&TestConfigMapHandler{})

	resultConfigMapHandler := GetConfigMapHandler()

	Assert(t, resultConfigMapHandler, NotNilVal())
}
