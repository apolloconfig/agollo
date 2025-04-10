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
	"fmt"
	"os"
	"sync"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/file"
)

var (
	// raw is the singleton instance of FileHandler for raw file operations
	raw file.FileHandler
	// rawOnce ensures thread-safe initialization of the raw FileHandler
	rawOnce sync.Once
)

// rawFileHandler extends FileHandler to write both raw content and namespace type
// when backing up configuration files. It provides functionality to store
// configurations in their original format alongside the JSON representation
type rawFileHandler struct {
	*FileHandler
}

// writeWithRaw writes the raw configuration content to a file
// Parameters:
//   - config: Apollo configuration containing the raw content
//   - configDir: Target directory for the configuration file
//
// Returns:
//   - error: Any error that occurred during the write operation
//
// This function:
// 1. Constructs the file path using namespace
// 2. Creates a new file
// 3. Writes the raw content if available
func writeWithRaw(config *config.ApolloConfig, configDir string) error {
	filePath := ""
	if configDir != "" {
		filePath = fmt.Sprintf("%s/%s", configDir, config.NamespaceName)
	} else {
		filePath = config.NamespaceName
	}

	file, e := os.Create(filePath)
	if e != nil {
		return e
	}
	defer file.Close()
	if config.Configurations["content"] != nil {
		_, e = file.WriteString(config.Configurations["content"].(string))
		if e != nil {
			return e
		}
	}
	return nil
}

// WriteConfigFile implements the FileHandler interface for raw file handling
// Parameters:
//   - config: Apollo configuration to be written
//   - configPath: Target directory path for the configuration file
//
// Returns:
//   - error: Any error that occurred during the write operation
//
// This method:
// 1. Creates the target directory if needed
// 2. Writes the raw content to a separate file
// 3. Writes the JSON format configuration
func (fileHandler *rawFileHandler) WriteConfigFile(config *config.ApolloConfig, configPath string) error {
	err := fileHandler.createDir(configPath)
	if err != nil {
		return err
	}

	err = writeWithRaw(config, configPath)
	if err != nil {
		log.Errorf("writeWithRaw fail! error:%v", err)
	}
	return jsonFileConfig.Write(config, fileHandler.GetConfigFile(configPath, config.AppID, config.NamespaceName))
}

// GetRawFileHandler returns a singleton instance of the raw file handler
// Returns:
//   - file.FileHandler: Thread-safe singleton instance of rawFileHandler
//
// This function uses sync.Once to ensure the handler is initialized only once
func GetRawFileHandler() file.FileHandler {
	rawOnce.Do(func() {
		raw = &rawFileHandler{}
	})
	return raw
}
