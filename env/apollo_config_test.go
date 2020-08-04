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

package env

import (
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/utils"
)

func TestSetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	config := &ApolloConnConfig{
		AppID:      "a",
		ReleaseKey: "releaseKey",
	}
	SetCurrentApolloConfig("a", config)
}

func TestGetCurrentApolloConfig(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	config := GetCurrentApolloConfig()["a"]
	Assert(t, config, NotNilVal())
	Assert(t, config.AppID, Equal("a"))
}

func TestGetCurrentApolloConfigReleaseKey(t *testing.T) {
	Assert(t, currentConnApolloConfig.configs, NotNilVal())
	key := GetCurrentApolloConfigReleaseKey("b")
	Assert(t, key, Equal(utils.Empty))

	key = GetCurrentApolloConfigReleaseKey("a")
	Assert(t, key, Equal("releaseKey"))
}

func TestApolloConfigInit(t *testing.T) {
	config := &ApolloConfig{}
	config.Init("appId", "cluster", "ns")

	Assert(t, config.AppID, Equal("appId"))
	Assert(t, config.Cluster, Equal("cluster"))
	Assert(t, config.NamespaceName, Equal("ns"))
}
