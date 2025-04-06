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

package parse

// ContentParser defines the interface for configuration content parsing
// This interface is used by Apollo to support different configuration formats
// (e.g., Properties, YAML, JSON) through format-specific implementations
type ContentParser interface {
	// Parse converts configuration content from its original format to a key-value map
	// Parameters:
	//   - configContent: The configuration content to parse, type depends on the format
	// Returns:
	//   - map[string]interface{}: Parsed configuration as key-value pairs
	//   - error: Any error that occurred during parsing
	// Implementations should handle format-specific parsing logic and
	// return a consistent map structure regardless of the input format
	Parse(configContent interface{}) (map[string]interface{}, error)
}
