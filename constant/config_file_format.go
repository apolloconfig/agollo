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

package constant

// ConfigFileFormat represents the supported configuration file formats in Apollo
// It is used to determine the appropriate parser for different configuration file types
type ConfigFileFormat string

const (
	// Properties represents Java properties file format (.properties)
	// This format is commonly used for storing key-value pairs in Java applications
	Properties ConfigFileFormat = ".properties"

	// XML represents XML file format (.xml)
	// Used for structured configuration data in XML format
	XML ConfigFileFormat = ".xml"

	// JSON represents JSON file format (.json)
	// Used for structured configuration data in JSON format
	JSON ConfigFileFormat = ".json"

	// YML represents YAML file format (.yml)
	// A human-readable format commonly used for configuration files
	YML ConfigFileFormat = ".yml"

	// YAML represents YAML file format (.yaml)
	// Alternative extension for YAML format files
	YAML ConfigFileFormat = ".yaml"

	// DEFAULT represents the default format (empty string)
	// Used when no specific format is specified or detected
	DEFAULT ConfigFileFormat = ""
)
