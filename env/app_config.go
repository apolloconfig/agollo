// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// appConfigFile defines the default configuration file name
	appConfigFile = "app.properties"
	// appConfigFilePath defines the environment variable name for custom config path
	appConfigFilePath = "AGOLLO_CONF"

	// defaultCluster defines the default cluster name
	defaultCluster = "default"
	// defaultNamespace defines the default namespace name
	defaultNamespace = "application"
)

var (
	// executeConfigFileOnce ensures thread-safe initialization of config file executor
	executeConfigFileOnce sync.Once
	// configFileExecutor handles configuration file operations
	configFileExecutor config.File
)

// InitFileConfig initializes configuration from the default properties file
// Returns:
//   - *config.AppConfig: Initialized configuration object, nil if initialization fails
//
// This function attempts to load configuration using default settings
func InitFileConfig() *config.AppConfig {
	// default use application.properties
	if initConfig, err := InitConfig(nil); err == nil {
		return initConfig
	}
	return nil
}

// InitConfig initializes configuration using a custom loader function
// Parameters:
//   - loadAppConfig: Custom function for loading application configuration
//
// Returns:
//   - *config.AppConfig: Initialized configuration object
//   - error: Any error that occurred during initialization
func InitConfig(loadAppConfig func() (*config.AppConfig, error)) (*config.AppConfig, error) {
	//init config file
	return getLoadAppConfig(loadAppConfig)
}

// getLoadAppConfig handles the actual configuration loading process
// Parameters:
//   - loadAppConfig: Optional custom function for loading application configuration
//
// Returns:
//   - *config.AppConfig: Loaded configuration object
//   - error: Any error that occurred during loading
//
// This function:
// 1. Uses custom loader if provided
// 2. Falls back to environment variable for config path
// 3. Uses default config file if no custom path is specified
func getLoadAppConfig(loadAppConfig func() (*config.AppConfig, error)) (*config.AppConfig, error) {
	if loadAppConfig != nil {
		return loadAppConfig()
	}
	configPath := os.Getenv(appConfigFilePath)
	if configPath == "" {
		configPath = appConfigFile
	}
	c, e := GetConfigFileExecutor().Load(configPath, Unmarshal)
	if c == nil {
		return nil, e
	}

	return c.(*config.AppConfig), e
}

// GetConfigFileExecutor returns a singleton instance of the configuration file executor
// Returns:
//   - config.File: Thread-safe instance of configuration file executor
//
// This function ensures the executor is initialized only once using sync.Once
func GetConfigFileExecutor() config.File {
	executeConfigFileOnce.Do(func() {
		configFileExecutor = &jsonConfig.ConfigFile{}
	})
	return configFileExecutor
}

// Unmarshal deserializes configuration data from bytes into AppConfig structure
// Parameters:
//   - b: Byte array containing configuration data
//
// Returns:
//   - interface{}: Unmarshaled configuration object
//   - error: Any error that occurred during unmarshaling
//
// This function:
// 1. Creates AppConfig with default values
// 2. Unmarshals JSON data into the config
// 3. Initializes the configuration
// 4. Returns the initialized config object
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
