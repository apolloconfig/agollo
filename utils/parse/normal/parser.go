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

package normal

// Parser represents the default content parser implementation
// It provides basic parsing functionality for Apollo configuration content
// This parser is used when no specific format parser is registered
type Parser struct {
}

// Parse converts configuration content to a key-value map
// Parameters:
//   - configContent: The configuration content to parse, can be of any type
//
// Returns:
//   - map[string]interface{}: Parsed configuration as key-value pairs
//   - error: Any error that occurred during parsing
//
// Note: This is a default implementation that returns nil values,
// intended to be used as a fallback when no specific parser is available
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}
