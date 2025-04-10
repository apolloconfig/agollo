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

package auth

// HTTPAuth defines the interface for HTTP authentication in Apollo client
// This interface provides methods for generating authentication headers
// for secure communication with Apollo configuration service
type HTTPAuth interface {
	// HTTPHeaders generates HTTP authentication headers for Apollo API requests
	// Parameters:
	//   - url: The target API endpoint URL
	//   - appID: Application identifier used for authentication
	//   - secret: Secret key used for generating signatures
	// Returns:
	//   - map[string][]string: A map of HTTP headers where:
	//     - key: HTTP header name (e.g., "Authorization", "Timestamp")
	//     - value: Array of header values
	// The implementation should generate all necessary headers for Apollo authentication,
	// including authorization signature and timestamp
	HTTPHeaders(url string, appID string, secret string) map[string][]string
}
