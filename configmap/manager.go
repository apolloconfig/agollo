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
	"encoding/json"
	"fmt"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sync"
	"time"
)

type K8sManager struct {
	clientSet    kubernetes.Interface
	k8sNamespace string
}

var (
	instance *K8sManager
	once     sync.Once
)

func GetK8sManager(k8sNamespace string) (*K8sManager, error) {
	if instance != nil {
		return instance, nil
	}

	once.Do(func() {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			instance = nil
			once = sync.Once{}
			log.Errorf("Error creating in-cluster inClusterConfig: %v", err)
			return
		}
		clientSet, err := kubernetes.NewForConfig(inClusterConfig)
		if err != nil {
			instance = nil
			once = sync.Once{}
			log.Errorf("Error creating Kubernetes client set: %v", err)
			return
		}
		instance = &K8sManager{
			clientSet:    clientSet,
			k8sNamespace: k8sNamespace,
		}
	})
	if instance == nil {
		return nil, fmt.Errorf("failed to create K8sManager instance")
	}
	return instance, nil
}

// SetConfigMapWithRetry 使用k8s版本号机制解决并发问题
func (m *K8sManager) SetConfigMapWithRetry(configMapName string, key string, config *config.ApolloConfig) error {
	var retryParam = wait.Backoff{
		Steps:    5,
		Duration: 10 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}

	err := retry.RetryOnConflict(retryParam, func() error {
		return m.SetConfigMap(configMapName, key, config)
	})

	return err
}

// SetConfigMap 将map[string]interface{}转换为JSON字符串，并创建或更新ConfigMap
func (m *K8sManager) SetConfigMap(configMapName string, key string, config *config.ApolloConfig) error {
	k8sNamespace := m.k8sNamespace

	jsonData, err := json.Marshal(config.Configurations)
	jsonString := string(jsonData)
	if err != nil {
		return fmt.Errorf("error marshaling data to JSON: %v", err)
	}
	log.Infof("Preparing Configmap content，JSON: %s", jsonString)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 尝试获取 ConfigMap，如果不存在则创建
	cm, err := m.clientSet.CoreV1().ConfigMaps(k8sNamespace).Get(ctx, configMapName, metaV1.GetOptions{})
	if errors.IsNotFound(err) {
		cm = &coreV1.ConfigMap{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      configMapName,
				Namespace: k8sNamespace,
			},
			Data: map[string]string{
				key: jsonString,
			},
		}

		_, err = m.clientSet.CoreV1().ConfigMaps(k8sNamespace).Create(ctx, cm, metaV1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error creating ConfigMap: %v", err)
		}
		log.Infof("ConfigMap %s created in namespace %s", configMapName, k8sNamespace)
	} else if err != nil {
		return fmt.Errorf("error getting ConfigMap: %v", err)
	} else {
		// ConfigMap 存在，更新数据
		cm.Data[key] = jsonString
		_, err = m.clientSet.CoreV1().ConfigMaps(k8sNamespace).Update(ctx, cm, metaV1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error updating ConfigMap: %v", err)
		}
		log.Infof("ConfigMap %s updated in namespace %s", configMapName, k8sNamespace)
	}
	return err
}

// GetConfigMap 从ConfigMap中获取JSON字符串，并反序列化为map[string]interface{}
func (m *K8sManager) GetConfigMap(configMapName string, key string) (map[string]interface{}, error) {
	k8sNamespace := m.k8sNamespace

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configMap, err := m.clientSet.CoreV1().ConfigMaps(k8sNamespace).Get(ctx, configMapName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ConfigMap: %v", err)
	}

	// 从ConfigMap中读取JSON数据
	jsonData, ok := configMap.Data[key]
	if !ok {
		return nil, fmt.Errorf("key: %v not found in ConfigMap", key)
	}

	var configurations map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &configurations)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to map[string]interface{}: %v", err)
	}

	return configurations, nil
}
