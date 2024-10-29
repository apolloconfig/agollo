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
	"context"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sync"
	"testing"
	"time"
)

func TestK8sManager_SetConfigMap(t *testing.T) {
	// 创建fake clientSet
	clientSet := fake.NewSimpleClientset()

	// 测试数据
	configMapName := "apollo-configcache-test-configmap"
	k8sNamespace := "default"
	key := "configKey"
	apolloConfig := &config.ApolloConfig{
		Configurations: map[string]interface{}{"key": "value"},
	}

	// 创建K8sManager实例
	manager := K8sManager{
		clientSet:    clientSet,
		k8sNamespace: k8sNamespace,
	}

	// 调用SetConfigMap方法
	err := manager.SetConfigMapWithRetry(configMapName, key, apolloConfig)
	assert.NoError(t, err)

	// 验证ConfigMap是否被创建
	configMap, err := clientSet.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), configMapName, metaV1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, configMap)
	assert.Equal(t, string(configMap.Data[key]),
		`{"key":"value"}`)
}

func TestK8sManager_GetConfigMap(t *testing.T) {
	// 创建fake clientSet，并预先创建一个ConfigMap
	clientSet := fake.NewSimpleClientset(&coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "apollo-configcache-test-configmap",
			Namespace: "default",
		},
		Data: map[string]string{
			"configKey": `{"key":"value"}`,
		},
	})

	// 创建K8sManager实例
	manager := K8sManager{
		clientSet:    clientSet,
		k8sNamespace: "default",
	}

	// 测试数据
	configMapName := "apollo-configcache-test-configmap"
	key := "configKey"

	// 调用GetConfigMap方法
	configurations, err := manager.GetConfigMap(configMapName, key)
	assert.NoError(t, err)
	assert.NotNil(t, configurations)
	assert.Equal(t, configurations["key"], "value")
}

func TestSetConfigMapWithRetryConcurrent(t *testing.T) {
	// 创建一个假的Kubernetes客户端
	clientSet := fake.NewSimpleClientset()

	// 创建K8sManager实例
	manager := &K8sManager{
		clientSet:    clientSet,
		k8sNamespace: "default",
	}

	// 定义测试数据
	configMapName := "apollo-configcache-test-configmap"
	k8sNamespace := "default"
	key := "test-key"
	configData1 := &config.ApolloConfig{
		Configurations: map[string]interface{}{
			"key1": "value1",
		},
	}
	configData2 := &config.ApolloConfig{
		Configurations: map[string]interface{}{
			"key2": "value2",
		},
	}

	// 首先创建一个ConfigMap
	_, err := clientSet.CoreV1().ConfigMaps(k8sNamespace).Create(context.Background(), &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      configMapName,
			Namespace: k8sNamespace,
		},
		Data: map[string]string{},
	}, metaV1.CreateOptions{})
	require.NoError(t, err)

	// 使用WaitGroup等待两个并发操作完成
	var wg sync.WaitGroup
	wg.Add(2)

	// 并发执行SetConfigMapWithRetry
	go func() {
		defer wg.Done()
		err := manager.SetConfigMapWithRetry(configMapName, key, configData1)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		// 为了模拟并发冲突，这里故意延迟一段时间后再执行
		time.Sleep(1 * time.Millisecond)
		err := manager.SetConfigMapWithRetry(configMapName, key, configData2)
		require.NoError(t, err)
	}()

	// 等待两个并发操作完成
	wg.Wait()

	// 验证ConfigMap是否更新成功
	cm, err := clientSet.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), configMapName, metaV1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, cm)

	// 这里只能验证其中一个更新操作成功的结果，因为两个操作是并发的，最终结果取决于哪个操作最后执行
	// 例如，这里假设configData2的更新是最后执行的
	require.Contains(t, cm.Data[key], `"key2":"value2"`)
}

func TestGetK8sManager_Singleton(t *testing.T) {
	// 重置全局变量
	instance = nil
	once = sync.Once{}

	// 模拟成功的 InClusterConfig 和 NewForConfig
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	})
	patches.ApplyFunc(kubernetes.NewForConfig, func(config *rest.Config) (*kubernetes.Clientset, error) {
		return &kubernetes.Clientset{}, nil
	})

	manager1, err1 := GetK8sManager("default")
	manager2, err2 := GetK8sManager("default")

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NotNil(t, manager1)
	assert.NotNil(t, manager2)
	assert.Equal(t, manager1, manager2, "GetK8sManager should return the same instance")
}

func TestGetK8sManager_InitSuccess(t *testing.T) {
	// 重置全局变量
	instance = nil
	once = sync.Once{}

	// 模拟成功的 InClusterConfig 和 NewForConfig
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	})
	patches.ApplyFunc(kubernetes.NewForConfig, func(config *rest.Config) (*kubernetes.Clientset, error) {
		return &kubernetes.Clientset{}, nil
	})

	manager, err := GetK8sManager("default")

	assert.Nil(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, "default", manager.k8sNamespace)
}
