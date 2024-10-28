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

package configmap

import (
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
)

var ApolloConfigCache = "apollo-configcache-"

type ConfigMapHandler struct {
	K8sManager *K8sManager
}

// WriteConfigFile write apollo config to configmap
func (c *ConfigMapHandler) WriteConfigFile(config *config.ApolloConfig, configPath string) error {
	configMapName := ApolloConfigCache + config.AppID
	key := config.Cluster + "-" + config.NamespaceName
	err := c.K8sManager.SetConfigMapWithRetry(configMapName, key, config)
	if err != nil {
		log.Errorf("Failed to write ConfigMap %s : %v", configMapName, err)
		return err
	}
	return nil
}

func (c *ConfigMapHandler) GetConfigFile(configDir string, appID string, namespace string) string {
	return ""
}

// LoadConfigFile load ApolloConfig from configmap
func (c *ConfigMapHandler) LoadConfigFile(configPath string, appID string, namespace string, cluster string) (*config.ApolloConfig, error) {
	var apolloConfig = &config.ApolloConfig{}
	var err error

	if cluster == "" {
		cluster = "default"
		log.Infof("cluster is empty, use default cluster")
	}
	configMapName := ApolloConfigCache + appID
	key := cluster + "-" + namespace
	// TODO 在这里把json转为ApolloConfig, ReleaseKey字段会丢失, 影响大不大
	apolloConfig.Configurations, err = c.K8sManager.GetConfigMap(configMapName, key)

	apolloConfig.AppID = appID
	apolloConfig.Cluster = cluster
	apolloConfig.NamespaceName = namespace
	return apolloConfig, err
}
