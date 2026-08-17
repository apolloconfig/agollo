// Copyright 2026 Apollo Authors
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

package agollo

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func parseYAML(content, format string) (map[string]interface{}, error) {
	var document map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return nil, fmt.Errorf("agollo: parse %s config: %w", format, err)
	}
	values := make(map[string]interface{})
	flattenYAML(values, "", document)
	return values, nil
}

func flattenYAML(values map[string]interface{}, prefix string, value interface{}) {
	switch typed := value.(type) {
	case map[interface{}]interface{}:
		for rawKey, child := range typed {
			key, ok := rawKey.(string)
			if !ok {
				continue
			}
			if prefix != "" {
				key = prefix + "." + key
			}
			flattenYAML(values, key, child)
		}
	case map[string]interface{}:
		for key, child := range typed {
			if prefix != "" {
				key = prefix + "." + key
			}
			flattenYAML(values, key, child)
		}
	default:
		if prefix != "" {
			values[prefix] = cloneValue(typed)
		}
	}
}
