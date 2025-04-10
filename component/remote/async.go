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
	"time"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/http"
	"github.com/apolloconfig/agollo/v4/utils"
)

const (
	// notifyConnectTimeout defines the timeout duration for notification connections
	// Default value is 10 minutes to maintain long polling connections
	notifyConnectTimeout = 10 * time.Minute

	// defaultContentKey is the key used to store configuration content
	defaultContentKey = "content"
)

// CreateAsyncApolloConfig creates and initializes a new asynchronous Apollo configuration instance
// Returns:
//   - ApolloConfig: A new instance of asyncApolloConfig
func CreateAsyncApolloConfig() ApolloConfig {
	a := &asyncApolloConfig{}
	a.remoteApollo = a
	return a
}

// asyncApolloConfig implements asynchronous configuration management for Apollo
// It extends AbsApolloConfig to provide async-specific functionality
type asyncApolloConfig struct {
	AbsApolloConfig
}

// GetNotifyURLSuffix constructs the URL suffix for notification API
// Parameters:
//   - notifications: JSON string containing notification information
//   - config: Application configuration
//
// Returns:
//   - string: The constructed URL suffix for notifications endpoint
func (*asyncApolloConfig) GetNotifyURLSuffix(notifications string, config config.AppConfig) string {
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}

// GetSyncURI constructs the URL for synchronizing configuration
// Parameters:
//   - config: Application configuration
//   - namespaceName: The namespace to sync
//
// Returns:
//   - string: The constructed URL for config sync endpoint
func (*asyncApolloConfig) GetSyncURI(config config.AppConfig, namespaceName string) string {
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s&label=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(config.GetCurrentApolloConfig().GetReleaseKey(namespaceName)),
		utils.GetInternal(),
		url.QueryEscape(config.Label))
}

// Sync synchronizes configurations from remote Apollo server
// Parameters:
//   - appConfigFunc: Function that provides application configuration
//
// Returns:
//   - []*config.ApolloConfig: Array of synchronized Apollo configurations
func (a *asyncApolloConfig) Sync(appConfigFunc func() config.AppConfig) []*config.ApolloConfig {
	appConfig := appConfigFunc()
	remoteConfigs, err := a.notifyRemoteConfig(appConfigFunc, utils.Empty)

	var apolloConfigs []*config.ApolloConfig
	if err != nil {
		apolloConfigs = loadBackupConfig(appConfig.NamespaceName, appConfig)
	}

	if len(remoteConfigs) == 0 || len(apolloConfigs) > 0 {
		return apolloConfigs
	}
	// just fetch the changed configurations, and update the namespace that has been fetched successfully
	for _, notifyConfig := range remoteConfigs {
		apolloConfig := a.SyncWithNamespace(notifyConfig.NamespaceName, appConfigFunc)
		if apolloConfig != nil {
			appConfig.GetNotificationsMap().UpdateNotify(notifyConfig.NamespaceName, notifyConfig.NotificationID)
			apolloConfigs = append(apolloConfigs, apolloConfig)
		}
	}
	return apolloConfigs
}

// CallBack creates a callback handler for HTTP requests
// Parameters:
//   - namespace: The namespace for which the callback is created
//
// Returns:
//   - http.CallBack: Callback structure with success and error handlers
func (*asyncApolloConfig) CallBack(namespace string) http.CallBack {
	return http.CallBack{
		SuccessCallBack:   createApolloConfigWithJSON,
		NotModifyCallBack: touchApolloConfigCache,
		Namespace:         namespace,
	}
}

// notifyRemoteConfig handles the long polling notification process
// Parameters:
//   - appConfigFunc: Function that provides application configuration
//   - namespace: The namespace to monitor
//
// Returns:
//   - []*config.Notification: Array of notifications
//   - error: Any error that occurred during the process
func (a *asyncApolloConfig) notifyRemoteConfig(appConfigFunc func() config.AppConfig, namespace string) ([]*config.Notification, error) {
	if appConfigFunc == nil {
		panic("can not find apollo config!please confirm!")
	}
	appConfig := appConfigFunc()
	notificationsMap := appConfig.GetNotificationsMap()
	urlSuffix := a.GetNotifyURLSuffix(notificationsMap.GetNotifies(namespace), appConfig)

	connectConfig := &env.ConnectConfig{
		URI:    urlSuffix,
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	connectConfig.Timeout = notifyConnectTimeout
	notifies, err := http.RequestRecovery(appConfig, connectConfig, &http.CallBack{
		SuccessCallBack: func(responseBody []byte, callback http.CallBack) (interface{}, error) {
			return toApolloConfig(responseBody)
		},
		NotModifyCallBack: touchApolloConfigCache,
		Namespace:         namespace,
	})

	if notifies == nil {
		return nil, err
	}

	return notifies.([]*config.Notification), err
}

// touchApolloConfigCache is a no-op function for cache touching operations
// Returns:
//   - error: Always returns nil
func touchApolloConfigCache() error {
	return nil
}

// toApolloConfig converts response body to Apollo notification array
// Parameters:
//   - resBody: Raw response body from Apollo server
//
// Returns:
//   - []*config.Notification: Parsed notification array
//   - error: Any error during parsing
func toApolloConfig(resBody []byte) ([]*config.Notification, error) {
	remoteConfig := make([]*config.Notification, 0)

	err := json.Unmarshal(resBody, &remoteConfig)

	if err != nil {
		log.Errorf("Unmarshal Msg Fail, error: %v", err)
		return nil, err
	}
	return remoteConfig, nil
}

// loadBackupConfig loads configuration from backup files when remote sync fails
// Parameters:
//   - namespace: The namespace to load
//   - appConfig: Application configuration
//
// Returns:
//   - []*config.ApolloConfig: Array of configurations loaded from backup
func loadBackupConfig(namespace string, appConfig config.AppConfig) []*config.ApolloConfig {
	apolloConfigs := make([]*config.ApolloConfig, 0)
	config.SplitNamespaces(namespace, func(namespace string) {
		c, err := extension.GetFileHandler().LoadConfigFile(appConfig.BackupConfigPath, appConfig.AppID, namespace)
		if err != nil {
			log.Errorf("LoadConfigFile error, error: %v", err)
			return
		}
		if c == nil {
			return
		}
		apolloConfigs = append(apolloConfigs, c)
	})
	return apolloConfigs
}

// createApolloConfigWithJSON creates Apollo configuration from JSON response
// Parameters:
//   - b: Raw JSON bytes
//   - callback: HTTP callback handler
//
// Returns:
//   - interface{}: Created Apollo configuration
//   - error: Any error during creation
//
// This function:
// 1. Unmarshals JSON into Apollo config
// 2. Determines appropriate parser based on namespace
// 3. Parses configuration content
// 4. Updates configuration map
func createApolloConfigWithJSON(b []byte, callback http.CallBack) (o interface{}, err error) {
	apolloConfig := &config.ApolloConfig{}
	err = json.Unmarshal(b, apolloConfig)
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

	content, ok := apolloConfig.Configurations[defaultContentKey]
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
