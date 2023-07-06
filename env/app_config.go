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
	"encoding/json"
	"os"
	"sync"

	"github.com/apolloconfig/agollo/v4/env/config"
	jsonConfig "github.com/apolloconfig/agollo/v4/env/config/json"
	"github.com/apolloconfig/agollo/v4/utils"
)

const (
	appConfigFile     = "app.properties"
	appConfigFilePath = "AGOLLO_CONF"

	defaultCluster   = "default"
	defaultNamespace = "application"
)

var (
	executeConfigFileOnce sync.Once
	configFileExecutor    config.File
)

type LoadAppConfigFunc func() (*config.AppConfig, error)

// LoadAppConfigFromFile 从配置文件读配置
func LoadAppConfigFromFile() (*config.AppConfig, error) {
	// 如果没有外部设定配置，从环境变量读取配置
	configPath := os.Getenv(appConfigFilePath)
	if configPath == "" {
		configPath = appConfigFile
	}
	conf, e := GetConfigFileExecutor().Load(configPath, Unmarshal)
	if conf == nil {
		return nil, e
	}
	return conf.(*config.AppConfig), e
}

// LoadAppConfig 加载配置
func LoadAppConfig(f LoadAppConfigFunc) (*config.AppConfig, error) {
	if f != nil {
		return f()
	}
	return LoadAppConfigFromFile()
}

// GetConfigFileExecutor 获取文件执行器
func GetConfigFileExecutor() config.File {
	executeConfigFileOnce.Do(func() {
		configFileExecutor = &jsonConfig.ConfigFile{}
	})
	return configFileExecutor
}

// Unmarshal 反序列化
func Unmarshal(b []byte) (interface{}, error) {
	appConfig := &config.AppConfig{
		Cluster:        defaultCluster,
		NamespaceName:  defaultNamespace,
		IsBackupConfig: true,
	}
	err := json.Unmarshal(b, appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}
	appConfig.Init()
	return appConfig, nil
}
