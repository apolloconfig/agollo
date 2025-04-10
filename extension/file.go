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

import "github.com/apolloconfig/agollo/v4/env/file"

// fileHandler is the global file handler instance for managing backup files
// It provides functionality for reading and writing Apollo configuration backups
var fileHandler file.FileHandler

// SetFileHandler sets the global file handler implementation
// Parameters:
//   - inFile: New file handler implementation to be used
//
// This function allows for custom file handling implementations to be injected
// into the Apollo client for different backup strategies or storage methods
func SetFileHandler(inFile file.FileHandler) {
	fileHandler = inFile
}

// GetFileHandler returns the current global file handler instance
// Returns:
//   - file.FileHandler: The current file handler implementation
//
// This function is used to obtain the file handler for backup operations
// such as reading or writing configuration files
func GetFileHandler() file.FileHandler {
	return fileHandler
}
