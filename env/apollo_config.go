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
	"sync"

	"github.com/zouyx/agollo/v3/utils"
)

var (
	currentConnApolloConfig = &currentApolloConfig{
		configs: make(map[string]*ApolloConnConfig, 1),
	}
)

type currentApolloConfig struct {
	l       sync.RWMutex
	configs map[string]*ApolloConnConfig
}

//SetCurrentApolloConfig 设置apollo配置
func SetCurrentApolloConfig(namespace string, connConfig *ApolloConnConfig) {
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.configs[namespace] = connConfig
}

//GetCurrentApolloConfig 获取Apollo链接配置
func GetCurrentApolloConfig() map[string]*ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.configs
}

//GetCurrentApolloConfigReleaseKey 获取release key
func GetCurrentApolloConfigReleaseKey(namespace string) string {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()
	config := currentConnApolloConfig.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

//ApolloConnConfig apollo链接配置
type ApolloConnConfig struct {
	AppID         string `json:"appId"`
	Cluster       string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey    string `json:"releaseKey"`
	sync.RWMutex
}

//ApolloConfig apollo配置
type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]interface{} `json:"configurations"`
}


//Init 初始化
func (a *ApolloConfig) Init(appID string, cluster string, namespace string) {
	a.AppID = appID
	a.Cluster = cluster
	a.NamespaceName = namespace
}

