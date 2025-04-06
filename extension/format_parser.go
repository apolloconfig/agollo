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

package extension

import (
	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/utils/parse"
)

// formatParser is a global map that stores content parsers for different configuration file formats
// Key: ConfigFileFormat - represents the format of the configuration file (e.g., JSON, YAML, Properties)
// Value: ContentParser - the corresponding parser implementation for that format
var formatParser = make(map[constant.ConfigFileFormat]parse.ContentParser, 0)

// AddFormatParser registers a new content parser for a specific configuration format
// Parameters:
//   - key: The format of the configuration file (e.g., JSON, YAML, Properties)
//   - contentParser: The parser implementation for the specified format
//
// This function enables support for parsing different configuration file formats
// by registering appropriate parser implementations
func AddFormatParser(key constant.ConfigFileFormat, contentParser parse.ContentParser) {
	formatParser[key] = contentParser
}

// GetFormatParser retrieves the content parser for a specific configuration format
// Parameters:
//   - key: The format of the configuration file to get the parser for
//
// Returns:
//   - parse.ContentParser: The parser implementation for the specified format
//     Returns nil if no parser is registered for the given format
func GetFormatParser(key constant.ConfigFileFormat) parse.ContentParser {
	return formatParser[key]
}
