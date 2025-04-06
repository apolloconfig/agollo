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

package roundrobin

import (
	"github.com/apolloconfig/agollo/v4/env/config"
)

// RoundRobin implements a simple round-robin load balancing strategy.
// This implementation currently provides a basic server selection mechanism
// that returns the first available server from the server list.
type RoundRobin struct {
}

// Load performs load balancing across the provided servers
// Parameters:
//   - servers: A map of server addresses to their corresponding ServerInfo objects
//
// Returns:
//   - *config.ServerInfo: The selected server for the current request
//     Returns nil if no available server is found
//
// Note: The current implementation selects the first available server
// that is not marked as down. A more sophisticated round-robin algorithm
// could be implemented to ensure better load distribution.
func (r *RoundRobin) Load(servers map[string]*config.ServerInfo) *config.ServerInfo {
	var returnServer *config.ServerInfo
	for _, server := range servers {
		// if some node has down then select next node
		if server.IsDown {
			continue
		}
		returnServer = server
		break
	}
	return returnServer
}
