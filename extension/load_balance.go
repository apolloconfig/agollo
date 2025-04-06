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

import "github.com/apolloconfig/agollo/v4/cluster"

// defaultLoadBalance is the global load balancer instance
// It provides functionality for distributing requests across multiple Apollo server nodes
var defaultLoadBalance cluster.LoadBalance

// SetLoadBalance sets the global load balancer implementation
// Parameters:
//   - loadBalance: New load balancer implementation to be used
//
// This function allows for custom load balancing strategies to be injected
// into the Apollo client for different server selection algorithms
func SetLoadBalance(loadBalance cluster.LoadBalance) {
	defaultLoadBalance = loadBalance
}

// GetLoadBalance returns the current global load balancer instance
// Returns:
//   - cluster.LoadBalance: The current load balancer implementation
//
// This function is used to obtain the load balancer for server selection
// during Apollo client operations
func GetLoadBalance() cluster.LoadBalance {
	return defaultLoadBalance
}
