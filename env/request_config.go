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

package env

import (
	"time"
)

// ConnectConfig defines the configuration for network requests to Apollo server
// This structure contains all necessary parameters for establishing and managing
// HTTP connections with the Apollo configuration service
type ConnectConfig struct {
	// Timeout specifies the maximum duration for HTTP requests
	// This value will be set to http.Client's timeout field
	// A zero or negative value means no timeout
	Timeout time.Duration

	// URI specifies the endpoint URL for the Apollo API
	// This should be the complete base URL for the Apollo server
	URI string

	// IsRetry indicates whether to retry failed requests
	// true: enable retry mechanism
	// false: disable retry mechanism
	IsRetry bool

	// AppID uniquely identifies the application in Apollo
	// This ID is used for authentication and configuration management
	AppID string

	// Secret is the authentication key for the application
	// Used for signing requests to ensure secure communication
	Secret string
}
