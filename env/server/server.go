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

package server

import (
	"strings"
	"sync"
	"time"

	"github.com/apolloconfig/agollo/v4/env/config"
)

// Global variables for managing Apollo server connections
var (
	// ipMap stores the mapping between config service IP and server information
	ipMap = make(map[string]*Info)
	// serverLock provides thread-safe access to ipMap
	serverLock sync.Mutex
	// nextTryConnectPeriod defines the waiting period (in seconds) before next connection attempt
	nextTryConnectPeriod int64 = 30
)

// Info represents Apollo server information and connection status
type Info struct {
	// serverMap stores the mapping of server URLs to their detailed information
	serverMap map[string]*config.ServerInfo
	// nextTryConnTime indicates the timestamp for the next connection attempt
	nextTryConnTime int64
}

// GetServers retrieves the server information map for a given configuration IP
// Parameters:
//   - configIp: The configuration service IP address
//
// Returns:
//   - map[string]*config.ServerInfo: Map of server information, nil if not found
func GetServers(configIp string) map[string]*config.ServerInfo {
	serverLock.Lock()
	defer serverLock.Unlock()
	if ipMap[configIp] == nil {
		return nil
	}
	return ipMap[configIp].serverMap
}

// GetServersLen returns the number of available servers for a given configuration IP
// Parameters:
//   - configIp: The configuration service IP address
//
// Returns:
//   - int: Number of available servers
func GetServersLen(configIp string) int {
	serverLock.Lock()
	defer serverLock.Unlock()
	s := ipMap[configIp]
	if s == nil || len(s.serverMap) == 0 {
		return 0
	}
	return len(s.serverMap)
}

// SetServers updates the server information for a given configuration IP
// Parameters:
//   - configIp: The configuration service IP address
//   - serverMap: New server information map to be set
func SetServers(configIp string, serverMap map[string]*config.ServerInfo) {
	serverLock.Lock()
	defer serverLock.Unlock()
	ipMap[configIp] = &Info{
		serverMap: serverMap,
	}
}

// SetDownNode marks a server node as unavailable
// Parameters:
//   - configService: The configuration service identifier
//   - serverHost: The host address of the server to be marked down
//
// This function:
// 1. Initializes server map if not exists
// 2. Updates next connection attempt time if needed
// 3. Marks specified server as down
func SetDownNode(configService string, serverHost string) {
	serverLock.Lock()
	defer serverLock.Unlock()
	s := ipMap[configService]
	if serverHost == "" {
		return
	}

	if s == nil || len(s.serverMap) == 0 {
		// init server map
		ipMap[configService] = &Info{
			serverMap: map[string]*config.ServerInfo{
				serverHost: {
					HomepageURL: serverHost,
				},
			},
		}
		s = ipMap[configService]
	}

	if serverHost == configService {
		s.nextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
	}

	for k, server := range s.serverMap {
		// if some node has down then select next node
		if strings.Contains(k, serverHost) {
			server.IsDown = true
		}
	}
}

// IsConnectDirectly determines whether to connect to the server directly
// Parameters:
//   - configIp: The configuration service IP address
//
// Returns:
//   - bool: true if should use meta server, false if should connect directly
//
// Note: The return value is inverse of the actual behavior for historical reasons
func IsConnectDirectly(configIp string) bool {
	serverLock.Lock()
	defer serverLock.Unlock()
	s := ipMap[configIp]
	if s == nil || len(s.serverMap) == 0 {
		return false
	}
	if s.nextTryConnTime >= 0 && s.nextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

// SetNextTryConnTime updates the next connection attempt time for a server
// Parameters:
//   - configIp: The configuration service IP address
//   - nextPeriod: Time period in seconds to wait before next attempt
//     If 0, uses default nextTryConnectPeriod
//
// This function ensures proper initialization of server info if not exists
func SetNextTryConnTime(configIp string, nextPeriod int64) {
	serverLock.Lock()
	defer serverLock.Unlock()
	s := ipMap[configIp]
	if s == nil || len(s.serverMap) == 0 {
		s = &Info{
			serverMap:       nil,
			nextTryConnTime: 0,
		}
		ipMap[configIp] = s
	}
	tmp := nextPeriod
	if tmp == 0 {
		tmp = nextTryConnectPeriod
	}
	s.nextTryConnTime = time.Now().Unix() + tmp
}
