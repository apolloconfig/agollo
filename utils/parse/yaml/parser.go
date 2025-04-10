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

package yaml

import (
	"bytes"

	"github.com/spf13/viper"

	"github.com/apolloconfig/agollo/v4/utils"
)

// vp is a global Viper instance for YAML parsing
// Using a single instance to improve performance
var vp = viper.New()

// init initializes the Viper instance with YAML configuration type
func init() {
	vp.SetConfigType("yaml")
}

// Parser implements the YAML format parser for Apollo configuration system
// It provides functionality to parse YAML format configuration content
// using the Viper library for robust YAML parsing capabilities
type Parser struct {
}

// Parse converts YAML format configuration content to a key-value map
// Parameters:
//   - configContent: The configuration content to parse, expected to be
//     a string containing YAML formatted data
//
// Returns:
//   - map[string]interface{}: Parsed configuration as key-value pairs
//   - error: Any error that occurred during parsing
//
// This method:
// 1. Validates and converts input to string
// 2. Uses Viper to parse YAML content
// 3. Converts parsed content to a flat key-value map
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if utils.Empty == content {
		return nil, nil
	}

	buffer := bytes.NewBufferString(content)
	// Use Viper to parse the YAML content
	err := vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	return convertToMap(vp), nil
}

// convertToMap converts Viper's parsed configuration to a flat key-value map
// Parameters:
//   - vp: Viper instance containing parsed configuration
//
// Returns:
//   - map[string]interface{}: Flattened key-value pairs from YAML configuration
//
// This function extracts all keys and their values from Viper's parsed
// configuration and stores them in a simple map structure
func convertToMap(vp *viper.Viper) map[string]interface{} {
	if vp == nil {
		return nil
	}

	m := make(map[string]interface{})
	for _, key := range vp.AllKeys() {
		m[key] = vp.Get(key)
	}
	return m
}
