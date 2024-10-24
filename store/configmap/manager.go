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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sync"
	"time"
)

// TODO 改为CAS更好，用版本号解决api server的并发 https://blog.csdn.net/boling_cavalry/article/details/128745382
type K8sManager struct {
	clientSet kubernetes.Interface
	mutex     sync.RWMutex // 添加读写锁
}

var (
	instance *K8sManager
	once     sync.Once
)

func GetK8sManager() (*K8sManager, error) {
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
			clientSet: clientSet,
			mutex:     sync.RWMutex{},
		}
	})
	if instance == nil {
		return nil, fmt.Errorf("failed to create K8sManager instance")
	}
	return instance, nil
}

// SetConfigMap 将map[string]interface{}转换为JSON字符串，并创建或更新ConfigMap
func (m *K8sManager) SetConfigMap(configMapName string, k8sNamespace string, key string, config *config.ApolloConfig) error {
	jsonData, err := json.Marshal(config.Configurations)
	jsonString := string(jsonData)
	if err != nil {
		return fmt.Errorf("error marshaling data to JSON: %v", err)
	}
	log.Infof("Preparing Configmap content，JSON: %s", jsonString)

	m.mutex.Lock() // 加锁
	defer m.mutex.Unlock()

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
func (m *K8sManager) GetConfigMap(configMapName string, k8sNamespace string, key string) (map[string]interface{}, error) {
	m.mutex.RLock() // 加读锁
	defer m.mutex.RUnlock()

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
