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

package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	jsonConfig "github.com/apolloconfig/agollo/v4/env/config/json"
)

// Suffix defines the default file extension for JSON configuration files
const Suffix = ".json"

var (
	// jsonFileConfig handles JSON format file operations for configurations
	jsonFileConfig = &jsonConfig.ConfigFile{}

	// configFileMap stores the mapping between namespace and file paths
	// Key: "{appID}-{namespace}", Value: full file path
	configFileMap = make(map[string]string, 1)

	// configFileMapLock provides thread-safe access to configFileMap
	configFileMapLock sync.Mutex

	// configFileDirMap caches directory existence status
	// Key: directory path, Value: whether directory exists
	configFileDirMap = make(map[string]bool, 1)

	// configFileDirMapLock provides thread-safe access to configFileDirMap
	configFileDirMapLock sync.Mutex
)

// FileHandler implements the default backup file operations for Apollo configurations
type FileHandler struct {
}

// createDir ensures the configuration directory exists
// Parameters:
//   - configPath: Target directory path to create
//
// Returns:
//   - error: Any error that occurred during directory creation
func (fileHandler *FileHandler) createDir(configPath string) error {
	if configPath == "" {
		return nil
	}

	configFileDirMapLock.Lock()
	defer configFileDirMapLock.Unlock()
	if !configFileDirMap[configPath] {
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Errorf("Create backup dir:%s fail, error:%v", configPath, err)
			return err
		}
		configFileDirMap[configPath] = true
	}
	return nil
}

// WriteConfigFile writes Apollo configuration to a JSON file
// Parameters:
//   - config: Apollo configuration to be written
//   - configPath: Target directory path for the configuration file
//
// Returns:
//   - error: Any error that occurred during the write operation
func (fileHandler *FileHandler) WriteConfigFile(config *config.ApolloConfig, configPath string) error {
	err := fileHandler.createDir(configPath)
	if err != nil {
		return err
	}
	return jsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.AppID, config.NamespaceName))
}

// GetConfigFile generates and caches the full configuration file path
// Parameters:
//   - configDir: Base directory for configuration files
//   - appID: Application identifier
//   - namespace: Configuration namespace
//
// Returns:
//   - string: Complete file path for the configuration file
//
// Thread-safe operation using mutex lock
func (fileHandler *FileHandler) GetConfigFile(configDir string, appID string, namespace string) string {
	key := fmt.Sprintf("%s-%s", appID, namespace)
	configFileMapLock.Lock()
	defer configFileMapLock.Unlock()
	fullPath := configFileMap[key]
	if fullPath == "" {
		filePath := fmt.Sprintf("%s%s", key, Suffix)
		if configDir != "" {
			configFileMap[namespace] = fmt.Sprintf("%s/%s", configDir, filePath)
		} else {
			configFileMap[namespace] = filePath
		}
	}
	return configFileMap[namespace]
}

// LoadConfigFile reads and parses an Apollo configuration from JSON file
// Parameters:
//   - configDir: Base directory for configuration files
//   - appID: Application identifier
//   - namespace: Configuration namespace
//
// Returns:
//   - *config.ApolloConfig: Parsed configuration object
//   - error: Any error that occurred during loading
//
// This method:
// 1. Constructs the file path
// 2. Reads the JSON file
// 3. Decodes the content into ApolloConfig structure
func (fileHandler *FileHandler) LoadConfigFile(configDir string, appID string, namespace string) (*config.ApolloConfig, error) {
	configFilePath := fileHandler.GetConfigFile(configDir, appID, namespace)
	log.Infof("load config file from: %s", configFilePath)
	c, e := jsonFileConfig.Load(configFilePath, func(b []byte) (interface{}, error) {
		config := &config.ApolloConfig{}
		e := json.NewDecoder(bytes.NewBuffer(b)).Decode(config)
		return config, e
	})

	if c == nil || e != nil {
		log.Errorf("loadConfigFile fail, error:%v", e)
		return nil, e
	}

	return c.(*config.ApolloConfig), e
}
