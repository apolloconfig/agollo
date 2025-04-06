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

package cluster

import (
	"github.com/apolloconfig/agollo/v4/env/config"
)

// LoadBalance defines the interface for load balancing strategies.
// Implementations of this interface can provide different algorithms
// for distributing load across multiple Apollo configuration servers.
type LoadBalance interface {
	// Load performs server selection based on the implemented load balancing strategy
	// Parameters:
	//   - servers: A map of server addresses to their corresponding ServerInfo objects
	//
	// Returns:
	//   - *config.ServerInfo: The selected server based on the load balancing algorithm
	//     Returns nil if no available server is found
	//
	// This method should handle server selection logic including:
	// - Checking server availability
	// - Implementing specific load balancing algorithms (e.g., round-robin, weighted random)
	// - Handling failure scenarios
	Load(servers map[string]*config.ServerInfo) *config.ServerInfo
}
