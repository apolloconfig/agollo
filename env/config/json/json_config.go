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
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/utils"
)

// ConfigFile implements JSON file read and write operations for Apollo configuration
// This type provides methods to load and save Apollo configurations in JSON format
type ConfigFile struct {
}

// Load reads and parses a JSON configuration file
// Parameters:
//   - fileName: The path to the JSON configuration file
//   - unmarshal: A function that defines how to unmarshal the file content
//
// Returns:
//   - interface{}: The parsed configuration object
//   - error: Any error that occurred during reading or parsing
func (t *ConfigFile) Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error) {
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.New("Fail to read config file:" + err.Error())
	}

	config, loadErr := unmarshal(fs)

	if utils.IsNotNil(loadErr) {
		return nil, errors.New("Load Json Config fail:" + loadErr.Error())
	}

	return config, nil
}

// Write saves Apollo configuration to a JSON file
// Parameters:
//   - content: The configuration content to be written
//   - configPath: The target file path where the configuration will be saved
//
// Returns:
//   - error: Any error that occurred during the write operation
//
// This method:
// 1. Validates the input content
// 2. Creates or overwrites the target file
// 3. Encodes the content as JSON
func (t *ConfigFile) Write(content interface{}, configPath string) error {
	if content == nil {
		log.Error("content is null can not write backup file")
		return errors.New("content is null can not write backup file")
	}
	file, e := os.Create(configPath)
	if e != nil {
		log.Errorf("writeConfigFile fail, error: %v", e)
		return e
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(content)
}
