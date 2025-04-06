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
	"github.com/apolloconfig/agollo/v4/protocol/auth"
)

// authSign is the global HTTP authentication handler instance
// It provides functionality for signing and authenticating requests
// to the Apollo configuration service
var authSign auth.HTTPAuth

// SetHTTPAuth sets the global HTTP authentication implementation
// Parameters:
//   - httpAuth: New HTTP authentication implementation to be used
//
// This function allows for custom authentication strategies to be injected
// into the Apollo client for different security requirements
func SetHTTPAuth(httpAuth auth.HTTPAuth) {
	authSign = httpAuth
}

// GetHTTPAuth returns the current global HTTP authentication instance
// Returns:
//   - auth.HTTPAuth: The current HTTP authentication implementation
//
// This function is used to obtain the authentication handler for
// securing requests to the Apollo configuration service
func GetHTTPAuth() auth.HTTPAuth {
	return authSign
}
