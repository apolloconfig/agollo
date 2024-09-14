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

import (
	"context"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestK8sManager_SetConfigMap(t *testing.T) {
	// 创建fake clientset
	clientSet := fake.NewSimpleClientset()

	// 创建K8sManager实例
	manager := K8sManager{clientSet: clientSet}

	// 测试数据
	configMapName := "test-configmap"
	configMapNamespace := "default"
	key := "configKey"
	apolloConfig := &config.ApolloConfig{
		Configurations: map[string]interface{}{"key": "value"},
	}

	// 调用SetConfigMap方法
	err := manager.SetConfigMap(configMapName, configMapNamespace, key, apolloConfig)
	assert.NoError(t, err)

	// 验证ConfigMap是否被创建
	configMap, err := clientSet.CoreV1().ConfigMaps(configMapNamespace).Get(context.Background(), configMapName, metaV1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, configMap)
	assert.Equal(t, string(configMap.Data[key]),
		`{"key":"value"}`)
}

func TestK8sManager_GetConfigMap(t *testing.T) {
	// 创建fake clientset，并预先创建一个ConfigMap
	clientSet := fake.NewSimpleClientset(&coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "default",
		},
		Data: map[string]string{
			"configKey": `{"key":"value"}`,
		},
	})

	// 创建K8sManager实例
	manager := K8sManager{clientSet: clientSet}

	// 测试数据
	configMapName := "test-configmap"
	configMapNamespace := "default"
	key := "configKey"

	// 调用GetConfigMap方法
	configurations, err := manager.GetConfigMap(configMapName, configMapNamespace, key)
	assert.NoError(t, err)
	assert.NotNil(t, configurations)
	assert.Equal(t, configurations["key"], "value")
}
