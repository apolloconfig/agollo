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
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sync"
	"time"
)

type K8sManager struct {
	clientSet kubernetes.Interface
}

var (
	instance *K8sManager
	once     sync.Once
)

func GetK8sManager() *K8sManager {
	once.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error()) // 处理错误
		}
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error()) // 处理错误
		}
		instance = &K8sManager{
			clientSet: clientSet,
		}
	})
	return instance
}

// SetConfigMap 将map[string]interface{}转换为JSON字符串，并创建或更新ConfigMap
func (m *K8sManager) SetConfigMap(configMapName string, configMapNamespace string, key string, config *config.ApolloConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 将ApolloConfig的配置信息数据转换为JSON字符串
	jsonData, err := json.Marshal(config.Configurations)
	if err != nil {
		return fmt.Errorf("error marshaling data to JSON: %v", err)
	}

	// 创建ConfigMap对象
	configMap := &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: map[string]string{
			key: string(jsonData),
		},
	}

	_, err = m.clientSet.CoreV1().ConfigMaps(configMapNamespace).Create(ctx, configMap, metaV1.CreateOptions{})
	if err != nil {
		_, err = m.clientSet.CoreV1().ConfigMaps(configMapNamespace).Update(ctx, configMap, metaV1.UpdateOptions{})
	}
	return err
}

// GetConfigMap 从ConfigMap中获取JSON字符串，并反序列化为map[string]interface{}
func (m *K8sManager) GetConfigMap(configMapName string, configMapNamespace string, key string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configMap, err := m.clientSet.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ConfigMap: %v", err)
	}

	// 从ConfigMap中读取JSON数据
	jsonData, ok := configMap.Data[key]
	if !ok {
		return nil, fmt.Errorf("key: %v not found in ConfigMap", key)
	}

	// 反序列化JSON配置信息到ApolloConfig的Configurations字段
	configTemp := &config.ApolloConfig{}
	err = json.Unmarshal([]byte(jsonData), &configTemp.Configurations)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to map[string]interface{}: %v", err)
	}

	return configTemp.Configurations, nil
}
