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

package agcache

// CacheInterface defines the contract for custom cache implementations
type CacheInterface interface {
	// Set stores a key-value pair in the cache with an expiration time
	// Parameters:
	//   - key: The unique identifier for the cache entry
	//   - value: The data to be stored
	//   - expireSeconds: Time in seconds after which the entry should expire
	// Returns:
	//   - error: Any error that occurred during the operation
	Set(key string, value interface{}, expireSeconds int) (err error)

	// EntryCount returns the total number of entries in the cache
	// Returns:
	//   - entryCount: The current number of entries stored in the cache
	EntryCount() (entryCount int64)

	// Get retrieves a value from the cache by its key
	// Parameters:
	//   - key: The unique identifier for the cache entry
	// Returns:
	//   - value: The stored value if found
	//   - error: Error if the key doesn't exist or any other error occurs
	Get(key string) (value interface{}, err error)

	// Del removes an entry from the cache by its key
	// Parameters:
	//   - key: The unique identifier of the entry to be deleted
	// Returns:
	//   - affected: True if the key was found and deleted, false otherwise
	Del(key string) (affected bool)

	// Range iterates over all key/value pairs in the cache
	// Parameters:
	//   - f: The function to be executed for each cache entry
	//        Return false from f to stop iteration
	Range(f func(key, value interface{}) bool)

	// Clear removes all entries from the cache
	Clear()
}

// CacheFactory defines the interface for creating cache instances
type CacheFactory interface {
	// Create instantiates and returns a new cache implementation
	// Returns:
	//   - CacheInterface: A new instance of a cache implementation
	Create() CacheInterface
}
