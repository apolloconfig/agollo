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
	"encoding/json"
	"fmt"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

const (
	appId              = "testAppId"
	configMapNamespace = "testConfigMapNamespace"
	cluster            = "testCluster"
	namespace          = "testNamespace"
)

var testData = map[string]interface{}{
	"stringKey": "stringValue",
	"intKey":    123,
	"boolKey":   true,
	"sliceKey":  []interface{}{1, 2, 3},
	"mapKey": map[string]interface{}{
		"nestedStringKey": "nestedStringValue",
		"nestedIntKey":    456,
	},
}

// TODO 更多边界的测试

func TestStore_LoadConfigMap(t *testing.T) {
	// 初始化fake clientset
	clientset := fake.NewSimpleClientset()
	jsonData, err := json.MarshalIndent(testData, "", "")
	if err != nil {
		fmt.Println("Error marshalling map to JSON:", err)
		return
	}

	// 创建一个ConfigMap对象
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appId,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			cluster + "+" + namespace: string(jsonData),
		},
	}

	// 使用fake clientset创建ConfigMap
	_, err = clientset.CoreV1().ConfigMaps(configMapNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	assert.NoError(t, err)

	// 初始化Store，注入fake clientset
	store := Store{
		K8sManager: &K8sManager{
			clientSet: clientset,
		},
	}

	var appConfig = config.AppConfig{
		AppID:         appId,
		NamespaceName: namespace,
		Cluster:       cluster,
	}

	// 执行
	loadedConfig, err := store.LoadConfigMap(appConfig, configMapNamespace)

	// 测试LoadConfigMap方法
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, testData, loadedConfig.Configurations)
}

func TestStore_WriteConfigMap(t *testing.T) {
	// 初始化fake clientset
	clientset := fake.NewSimpleClientset()
	jsonData, err := json.MarshalIndent(testData, "", "")
	if err != nil {
		fmt.Println("Error marshalling map to JSON:", err)
		return
	}

	// 初始化Store，注入fake clientset
	store := Store{
		K8sManager: &K8sManager{
			clientSet: clientset,
		},
	}

	// 反序列化到ApolloConfig
	apolloConfig := &config.ApolloConfig{}
	err = json.Unmarshal(jsonData, apolloConfig)
	assert.NoError(t, err)
	apolloConfig.AppID = appId
	apolloConfig.Cluster = cluster
	apolloConfig.NamespaceName = namespace

	// 测试WriteConfigMap方法
	err = store.WriteConfigMap(apolloConfig, configMapNamespace)
	assert.NoError(t, err)

	// 验证ConfigMap是否被正确创建或更新
	configMap, err := clientset.CoreV1().ConfigMaps(configMapNamespace).Get(context.TODO(), appId, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, configMap)
	assert.Equal(t, testData, configMap.Data)

}
