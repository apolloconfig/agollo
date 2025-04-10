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

package remote

import (
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/protocol/http"
)

// ApolloConfig defines the interface for interacting with Apollo Configuration Center
// This interface provides methods for both synchronous and asynchronous configuration updates
type ApolloConfig interface {
	// GetNotifyURLSuffix constructs the URL suffix for long polling notifications
	// Parameters:
	//   - notifications: JSON string containing notification information
	//   - config: Application configuration instance
	// Returns:
	//   - string: The constructed URL suffix for notifications endpoint
	GetNotifyURLSuffix(notifications string, config config.AppConfig) string

	// GetSyncURI constructs the URL for synchronizing configuration
	// Parameters:
	//   - config: Application configuration instance
	//   - namespaceName: The namespace to synchronize
	// Returns:
	//   - string: The constructed URL for configuration synchronization
	GetSyncURI(config config.AppConfig, namespaceName string) string

	// Sync synchronizes all configurations from Apollo server
	// Parameters:
	//   - appConfigFunc: Function that provides application configuration
	// Returns:
	//   - []*config.ApolloConfig: Array of synchronized Apollo configurations
	Sync(appConfigFunc func() config.AppConfig) []*config.ApolloConfig

	// CallBack creates a callback handler for specific namespace
	// Parameters:
	//   - namespace: The namespace for which the callback is created
	// Returns:
	//   - http.CallBack: Callback structure with success and error handlers
	CallBack(namespace string) http.CallBack

	// SyncWithNamespace synchronizes configuration for a specific namespace
	// Parameters:
	//   - namespace: The namespace to synchronize
	//   - appConfigFunc: Function that provides application configuration
	// Returns:
	//   - *config.ApolloConfig: The synchronized configuration for the specified namespace
	SyncWithNamespace(namespace string, appConfigFunc func() config.AppConfig) *config.ApolloConfig
}
