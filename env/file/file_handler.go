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

package file

import (
	"github.com/apolloconfig/agollo/v4/env/config"
)

// FileHandler defines the interface for handling Apollo configuration file operations
// This interface provides methods for reading, writing, and managing configuration files
type FileHandler interface {
	// WriteConfigFile writes Apollo configuration to a backup file
	// Parameters:
	//   - config: Apollo configuration to be written
	//   - configPath: Target path for the configuration file
	// Returns:
	//   - error: Any error that occurred during the write operation
	WriteConfigFile(config *config.ApolloConfig, configPath string) error

	// GetConfigFile constructs the configuration file path
	// Parameters:
	//   - configDir: Base directory for configuration files
	//   - appID: Application identifier
	//   - namespace: Configuration namespace
	// Returns:
	//   - string: Complete file path for the configuration file
	GetConfigFile(configDir string, appID string, namespace string) string

	// LoadConfigFile reads and parses an Apollo configuration file
	// Parameters:
	//   - configDir: Base directory for configuration files
	//   - appID: Application identifier
	//   - namespace: Configuration namespace
	// Returns:
	//   - *config.ApolloConfig: Parsed configuration object
	//   - error: Any error that occurred during loading
	LoadConfigFile(configDir string, appID string, namespace string) (*config.ApolloConfig, error)
}
