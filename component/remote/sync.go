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
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/http"
	"github.com/apolloconfig/agollo/v4/utils"
)

// CreateSyncApolloConfig creates a new instance of synchronous Apollo configuration client
// Returns:
//   - ApolloConfig: A new instance of syncApolloConfig
func CreateSyncApolloConfig() ApolloConfig {
	a := &syncApolloConfig{}
	a.remoteApollo = a
	return a
}

// syncApolloConfig implements synchronous configuration management for Apollo
// It extends AbsApolloConfig to provide sync-specific functionality
type syncApolloConfig struct {
	AbsApolloConfig
}

// GetNotifyURLSuffix returns an empty string as notifications are not used in sync mode
// This method is implemented to satisfy the ApolloConfig interface
func (*syncApolloConfig) GetNotifyURLSuffix(notifications string, config config.AppConfig) string {
	return ""
}

// GetSyncURI constructs the URL for synchronous configuration retrieval
// Parameters:
//   - config: Application configuration instance
//   - namespaceName: The namespace to retrieve
//
// Returns:
//   - string: The constructed URL for fetching JSON configuration files
func (*syncApolloConfig) GetSyncURI(config config.AppConfig, namespaceName string) string {
	return fmt.Sprintf("configfiles/json/%s/%s/%s?&ip=%s&label=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		utils.GetInternal(),
		url.QueryEscape(config.Label))
}

// CallBack creates a callback handler for processing JSON configuration files
// Parameters:
//   - namespace: The namespace for which the callback is created
//
// Returns:
//   - http.CallBack: Callback structure with JSON processing handlers
func (*syncApolloConfig) CallBack(namespace string) http.CallBack {
	return http.CallBack{
		SuccessCallBack:   processJSONFiles,
		NotModifyCallBack: touchApolloConfigCache,
		Namespace:         namespace,
	}
}

// processJSONFiles processes the JSON response from Apollo server
// Parameters:
//   - b: Raw JSON bytes from the response
//   - callback: Callback containing namespace information
//
// Returns:
//   - interface{}: Processed Apollo configuration
//   - error: Any error during processing
//
// This function:
// 1. Creates a new Apollo configuration instance
// 2. Unmarshals JSON into configurations
// 3. Applies appropriate format parser based on namespace
// 4. Processes configuration content
func processJSONFiles(b []byte, callback http.CallBack) (o interface{}, err error) {
	apolloConfig := &config.ApolloConfig{}
	apolloConfig.NamespaceName = callback.Namespace

	configurations := make(map[string]interface{}, 0)
	apolloConfig.Configurations = configurations
	err = json.Unmarshal(b, &apolloConfig.Configurations)

	if utils.IsNotNil(err) {
		return nil, err
	}

	parser := extension.GetFormatParser(constant.ConfigFileFormat(path.Ext(apolloConfig.NamespaceName)))
	if parser == nil {
		parser = extension.GetFormatParser(constant.DEFAULT)
	}

	if parser == nil {
		return apolloConfig, nil
	}

	content, ok := configurations[defaultContentKey]
	if !ok {
		content = string(b)
	}
	m, err := parser.Parse(content)
	if err != nil {
		log.Debugf("GetContent fail! error: %v", err)
	}

	if len(m) > 0 {
		apolloConfig.Configurations = m
	}
	return apolloConfig, nil
}

// Sync synchronizes configurations for all namespaces
// Parameters:
//   - appConfigFunc: Function that provides application configuration
//
// Returns:
//   - []*config.ApolloConfig: Array of synchronized configurations
//
// This method:
// 1. Retrieves configurations for each namespace
// 2. Falls back to backup configurations if sync fails
// 3. Aggregates all configurations into a single array
func (a *syncApolloConfig) Sync(appConfigFunc func() config.AppConfig) []*config.ApolloConfig {
	appConfig := appConfigFunc()
	configs := make([]*config.ApolloConfig, 0, 8)
	config.SplitNamespaces(appConfig.NamespaceName, func(namespace string) {
		apolloConfig := a.SyncWithNamespace(namespace, appConfigFunc)
		if apolloConfig != nil {
			configs = append(configs, apolloConfig)
			return
		}
		configs = append(configs, loadBackupConfig(namespace, appConfig)...)
	})
	return configs
}
