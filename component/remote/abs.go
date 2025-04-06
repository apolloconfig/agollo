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

package remote

import (
	"time"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/protocol/http"
)

// AbsApolloConfig represents an abstract Apollo configuration handler
// It provides base functionality for interacting with Apollo configuration server
type AbsApolloConfig struct {
	// remoteApollo is the interface for remote Apollo operations
	remoteApollo ApolloConfig
}

// SyncWithNamespace synchronizes configuration for a specific namespace from Apollo server
// Parameters:
//   - namespace: The configuration namespace to sync
//   - appConfigFunc: Function that provides the application configuration
//
// Returns:
//   - *config.ApolloConfig: The synchronized configuration, or nil if sync fails
//
// This method:
// 1. Validates the input parameters
// 2. Constructs the connection configuration
// 3. Makes HTTP request to Apollo server
// 4. Handles any errors during synchronization
func (a *AbsApolloConfig) SyncWithNamespace(namespace string, appConfigFunc func() config.AppConfig) *config.ApolloConfig {
	if appConfigFunc == nil {
		panic("can not find apollo config!please confirm!")
	}
	appConfig := appConfigFunc()
	urlSuffix := a.remoteApollo.GetSyncURI(appConfig, namespace)

	// Configure connection parameters for Apollo server
	c := &env.ConnectConfig{
		URI:     urlSuffix,
		AppID:   appConfig.AppID,
		Secret:  appConfig.Secret,
		Timeout: notifyConnectTimeout,
		IsRetry: true,
	}
	// Override timeout if specified in application config
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}

	// Execute synchronization request with error recovery
	callback := a.remoteApollo.CallBack(namespace)
	apolloConfig, err := http.RequestRecovery(appConfig, c, &callback)
	if err != nil {
		log.Errorf("request %s fail, error:%v", urlSuffix, err)
		return nil
	}

	if apolloConfig == nil {
		log.Debug("apolloConfig is nil")
		return nil
	}

	return apolloConfig.(*config.ApolloConfig)
}
