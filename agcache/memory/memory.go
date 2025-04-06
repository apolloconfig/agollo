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

package memory

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/apolloconfig/agollo/v4/agcache"
)

// DefaultCache implements a thread-safe in-memory cache using sync.Map
type DefaultCache struct {
	defaultCache sync.Map // The underlying thread-safe map for storing cache entries
	count        int64    // Counter for tracking the number of cache entries
}

// Set stores a key-value pair in the cache
// Parameters:
//   - key: The unique identifier for the cache entry
//   - value: The data to be stored
//   - expireSeconds: Time in seconds after which the entry should expire (currently not implemented)
//
// Returns:
//   - error: Always returns nil as the operation cannot fail
func (d *DefaultCache) Set(key string, value interface{}, expireSeconds int) (err error) {
	d.defaultCache.Store(key, value)
	atomic.AddInt64(&d.count, int64(1))
	return nil
}

// EntryCount returns the total number of entries in the cache
// Returns:
//   - entryCount: The current number of entries stored in the cache
func (d *DefaultCache) EntryCount() (entryCount int64) {
	c := atomic.LoadInt64(&d.count)
	return c
}

// Get retrieves a value from the cache by its key
// Parameters:
//   - key: The unique identifier for the cache entry
//
// Returns:
//   - value: The stored value if found
//   - error: Error if the key doesn't exist in the cache
func (d *DefaultCache) Get(key string) (value interface{}, err error) {
	v, ok := d.defaultCache.Load(key)
	if !ok {
		return nil, errors.New("load default cache fail")
	}
	return v, nil
}

// Range iterates over all key/value pairs in the cache
// Parameters:
//   - f: The function to be executed for each cache entry
//     Return false from f to stop iteration
func (d *DefaultCache) Range(f func(key, value interface{}) bool) {
	d.defaultCache.Range(f)
}

// Del removes an entry from the cache by its key
// Parameters:
//   - key: The unique identifier of the entry to be deleted
//
// Returns:
//   - affected: Always returns true regardless of whether the key existed
func (d *DefaultCache) Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	atomic.AddInt64(&d.count, int64(-1))
	return true
}

// Clear removes all entries from the cache
// This operation reinitializes the underlying sync.Map and resets the counter
func (d *DefaultCache) Clear() {
	d.defaultCache = sync.Map{}
	atomic.StoreInt64(&d.count, int64(0))
}

// DefaultCacheFactory is a factory for creating new instances of DefaultCache
type DefaultCacheFactory struct {
}

// Create instantiates and returns a new DefaultCache instance
// Returns:
//   - agcache.CacheInterface: A new instance of DefaultCache
func (d *DefaultCacheFactory) Create() agcache.CacheInterface {
	return &DefaultCache{}
}
