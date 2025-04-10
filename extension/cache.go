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

import "github.com/apolloconfig/agollo/v4/agcache"

var (
	// globalCacheFactory is the singleton instance of the cache factory
	// It provides a centralized way to create and manage cache instances
	globalCacheFactory agcache.CacheFactory
)

// GetCacheFactory returns the global cache factory instance
// Returns:
//   - agcache.CacheFactory: The current cache factory implementation
//
// This function is used to obtain the cache factory for creating new cache instances
func GetCacheFactory() agcache.CacheFactory {
	return globalCacheFactory
}

// SetCacheFactory updates the global cache factory implementation
// Parameters:
//   - cacheFactory: New cache factory implementation to be used
//
// This function allows for custom cache implementations to be injected
// into the Apollo client for different caching strategies
func SetCacheFactory(cacheFactory agcache.CacheFactory) {
	globalCacheFactory = cacheFactory
}
