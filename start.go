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

package agollo

import (
	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/cluster"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/file"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/auth"
)

// SetSignature configures a custom HTTP authentication component
// Parameters:
//   - auth: Custom implementation of HTTPAuth interface for request authentication
//
// This function allows users to:
// 1. Override the default authentication mechanism
// 2. Implement custom authentication logic for Apollo server requests
// Note: If auth is nil, the function will have no effect
func SetSignature(auth auth.HTTPAuth) {
	if auth != nil {
		extension.SetHTTPAuth(auth)
	}
}

// SetBackupFileHandler configures a custom backup file handler component
// Parameters:
//   - file: Custom implementation of FileHandler interface for backup operations
//
// This function enables:
// 1. Custom backup file storage mechanisms
// 2. Custom backup file formats
// 3. Custom backup locations
// Note: If file is nil, the function will have no effect
func SetBackupFileHandler(file file.FileHandler) {
	if file != nil {
		extension.SetFileHandler(file)
	}
}

// SetLoadBalance configures a custom load balancing component
// Parameters:
//   - loadBalance: Custom implementation of LoadBalance interface
//
// This function allows:
// 1. Custom server selection strategies
// 2. Custom load balancing algorithms
// 3. Custom health check mechanisms
// Note: If loadBalance is nil, the function will have no effect
func SetLoadBalance(loadBalance cluster.LoadBalance) {
	if loadBalance != nil {
		extension.SetLoadBalance(loadBalance)
	}
}

// SetLogger configures a custom logging component
// Parameters:
//   - loggerInterface: Custom implementation of LoggerInterface
//
// This function enables:
// 1. Custom log formatting
// 2. Custom log levels
// 3. Custom log output destinations
// Note: If loggerInterface is nil, the function will have no effect
func SetLogger(loggerInterface log.LoggerInterface) {
	if loggerInterface != nil {
		log.InitLogger(loggerInterface)
	}
}

// SetCache configures a custom cache component
// Parameters:
//   - cacheFactory: Custom implementation of CacheFactory interface
//
// This function allows:
// 1. Custom cache storage mechanisms
// 2. Custom cache eviction policies
// 3. Custom cache serialization methods
// Note: If cacheFactory is nil, the function will have no effect
func SetCache(cacheFactory agcache.CacheFactory) {
	if cacheFactory != nil {
		extension.SetCacheFactory(cacheFactory)
	}
}
