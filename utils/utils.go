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

package utils

import (
	"net"
	"os"
	"reflect"
	"sync"
)

const (
	// Empty represents an empty string constant
	// Used throughout the application for string comparisons and defaults
	Empty = ""
)

var (
	// internalIPOnce ensures the internal IP is retrieved only once
	internalIPOnce sync.Once
	// internalIP stores the cached internal IP address of the machine
	internalIP = ""
)

// GetInternal retrieves the internal IPv4 address of the machine
// Returns:
//   - string: The first non-loopback IPv4 address found
//
// This function:
// 1. Uses sync.Once to ensure single execution
// 2. Retrieves all network interface addresses
// 3. Finds the first non-loopback IPv4 address
// 4. Caches the result for subsequent calls
func GetInternal() string {
	internalIPOnce.Do(func() {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			os.Stderr.WriteString("Oops:" + err.Error())
			os.Exit(1)
		}
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					internalIP = ipnet.IP.To4().String()
				}
			}
		}
	})
	return internalIP
}

// IsNotNil checks if an object is not nil
// Parameters:
//   - object: The interface{} to check
//
// Returns:
//   - bool: true if the object is not nil, false otherwise
//
// This is a convenience wrapper around IsNilObject with inverted logic
func IsNotNil(object interface{}) bool {
	return !IsNilObject(object)
}

// IsNilObject determines if an object is nil or effectively nil
// Parameters:
//   - object: The interface{} to check
//
// Returns:
//   - bool: true if the object is nil or a nil interface value
//
// This function performs a thorough nil check that handles:
// 1. Direct nil values
// 2. nil interface values (chan, func, interface, map, pointer, slice)
// 3. Zero-value interfaces of reference types
func IsNilObject(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}
