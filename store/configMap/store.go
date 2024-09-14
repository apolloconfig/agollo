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

package configMap

import "github.com/apolloconfig/agollo/v4/env/config"

// TODO 并发加锁

type Store struct {
	// 模块化依赖注入，便于测试
	K8sManager *K8sManager
}

// LoadConfigMap load ApolloConfig from configmap
func (c *Store) LoadConfigMap(config *config.ApolloConfig, configMapNamespace string) (*config.ApolloConfig, error) {
	configMapName := config.AppID
	key := config.Cluster + "+" + config.NamespaceName
	// TODO 在这里把json转为ApolloConfig, 但ReleaseKey字段会丢失
	config.Configurations, _ = c.K8sManager.GetConfigMap(configMapName, configMapNamespace, key)
	return config, nil
}

// WriteConfigMap write apollo config to configmap
func (c *Store) WriteConfigMap(config *config.ApolloConfig, configMapNamespace string) error {
	// AppId作为configMap的name,cluster+namespace作为key, value为config
	configMapName := config.AppID
	key := config.Cluster + "+" + config.NamespaceName
	return c.K8sManager.SetConfigMap(configMapName, configMapNamespace, key, config)
}
