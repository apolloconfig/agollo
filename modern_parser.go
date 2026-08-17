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
	"bytes"
	"fmt"

	"github.com/spf13/viper"
)

func parseYAML(content, format string) (map[string]interface{}, error) {
	parser := viper.New()
	parser.SetConfigType(format)
	if err := parser.ReadConfig(bytes.NewBufferString(content)); err != nil {
		return nil, fmt.Errorf("agollo: parse %s config: %w", format, err)
	}
	values := make(map[string]interface{})
	for _, key := range parser.AllKeys() {
		values[key] = cloneValue(parser.Get(key))
	}
	return values, nil
}
