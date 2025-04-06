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

package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/apolloconfig/agollo/v4/utils"
)

var (
	// defaultNotificationID represents the default notification ID for new configurations
	defaultNotificationID = int64(-1)
	// Comma is the delimiter used for splitting namespace strings
	Comma = ","
)

// File defines the interface for reading and writing configuration files
type File interface {
	// Load reads and unmarshals a configuration file
	// Parameters:
	//   - fileName: Path to the configuration file
	//   - unmarshal: Function to unmarshal the file content
	// Returns:
	//   - interface{}: Unmarshaled configuration
	//   - error: Any error that occurred during loading
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	// Write saves configuration content to a file
	// Parameters:
	//   - content: Configuration content to write
	//   - configPath: Target file path
	// Returns:
	//   - error: Any error that occurred during writing
	Write(content interface{}, configPath string) error
}

// AppConfig represents the application configuration for Apollo client
type AppConfig struct {
	AppID             string `json:"appId"`                         // Unique identifier for the application
	Cluster           string `json:"cluster"`                       // Cluster name (e.g., dev, prod)
	NamespaceName     string `json:"namespaceName"`                 // Namespace for configuration isolation
	IP                string `json:"ip"`                            // Apollo server IP/host
	IsBackupConfig    bool   `default:"true" json:"isBackupConfig"` // Whether to backup configurations
	BackupConfigPath  string `json:"backupConfigPath"`              // Path for backup configuration files
	Secret            string `json:"secret"`                        // Authentication secret
	Label             string `json:"label"`                         // Configuration label
	SyncServerTimeout int    `json:"syncServerTimeout"`             // Timeout for server synchronization
	// MustStart controls whether the first sync must succeed
	MustStart               bool                 `default:"false"`
	notificationsMap        *notificationsMap    // Manages configuration change notifications
	currentConnApolloConfig *CurrentApolloConfig // Current Apollo connection configuration
}

// ServerInfo contains information about an Apollo server instance
type ServerInfo struct {
	AppName     string `json:"appName"`     // Name of the Apollo application
	InstanceID  string `json:"instanceId"`  // Unique identifier for the server instance
	HomepageURL string `json:"homepageUrl"` // Base URL of the server
	IsDown      bool   `json:"-"`           // Indicates if the server is unavailable
}

// GetIsBackupConfig returns whether to backup configuration after fetching from Apollo
// Returns:
//   - bool: true if backup is enabled (default), false otherwise
func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

// GetBackupConfigPath returns the path where configuration backups are stored
// Returns:
//   - string: The configured backup path
func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
}

// GetHost returns the Apollo server host URL with proper formatting
// Returns:
//   - string: The formatted host URL, ensuring it ends with a forward slash
func (a *AppConfig) GetHost() string {
	u, err := url.Parse(a.IP)
	if err != nil {
		return a.IP
	}
	if !strings.HasSuffix(u.Path, "/") {
		return u.String() + "/"
	}
	return u.String()
}

// Init initializes the AppConfig instance
// This method sets up the current Apollo configuration and notifications map
func (a *AppConfig) Init() {
	a.currentConnApolloConfig = CreateCurrentApolloConfig()
	a.initAllNotifications(nil)
}

// Notification represents a configuration change notification from Apollo
type Notification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationID int64  `json:"notificationId"`
}

// initAllNotifications initializes the notifications map for all configured namespaces
// Parameters:
//   - callback: Optional function to be executed for each namespace during initialization
//     The callback can be used for custom processing of each namespace
func (a *AppConfig) initAllNotifications(callback func(namespace string)) {
	ns := SplitNamespaces(a.NamespaceName, callback)
	a.notificationsMap = &notificationsMap{
		notifications: ns,
	}
}

// SplitNamespaces splits a comma-separated namespace string and initializes notifications
// Parameters:
//   - namespacesStr: Comma-separated string of namespace names
//   - callback: Optional function to be executed for each namespace after splitting
//
// Returns:
//   - sync.Map: Thread-safe map containing namespace names as keys and defaultNotificationID as values
//
// This function:
// 1. Splits the input string by comma
// 2. Processes each namespace individually
// 3. Initializes each namespace with default notification ID
func SplitNamespaces(namespacesStr string, callback func(namespace string)) sync.Map {
	namespaces := sync.Map{}
	split := strings.Split(namespacesStr, Comma)
	for _, namespace := range split {
		if callback != nil {
			callback(namespace)
		}
		namespaces.Store(namespace, defaultNotificationID)
	}
	return namespaces
}

// GetNotificationsMap returns the notifications map for the application
// Returns:
//   - *notificationsMap: Thread-safe map containing all namespace notifications
//
// This map is used to track configuration changes across different namespaces
func (a *AppConfig) GetNotificationsMap() *notificationsMap {
	return a.notificationsMap
}

// GetServicesConfigURL constructs the URL for fetching service configurations
// Returns:
//   - string: Complete URL with proper encoding for accessing Apollo configuration services
//
// The URL includes:
// 1. Base host URL
// 2. Application ID
// 3. Client IP address
func (a *AppConfig) GetServicesConfigURL() string {
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		a.GetHost(),
		url.QueryEscape(a.AppID),
		utils.GetInternal())
}

// SetCurrentApolloConfig nolint
func (a *AppConfig) SetCurrentApolloConfig(apolloConfig *ApolloConnConfig) {
	a.currentConnApolloConfig.Set(apolloConfig.NamespaceName, apolloConfig)
}

// GetCurrentApolloConfig nolint
func (a *AppConfig) GetCurrentApolloConfig() *CurrentApolloConfig {
	return a.currentConnApolloConfig
}

// map[string]int64
type notificationsMap struct {
	notifications sync.Map // Thread-safe map storing notification IDs by namespace
}

// UpdateAllNotifications updates notification IDs for multiple configurations
// Parameters:
//   - remoteConfigs: Array of notifications from the remote server
func (n *notificationsMap) UpdateAllNotifications(remoteConfigs []*Notification) {
	for _, remoteConfig := range remoteConfigs {
		if remoteConfig.NamespaceName == "" {
			continue
		}
		if n.GetNotify(remoteConfig.NamespaceName) == 0 {
			continue
		}

		n.setNotify(remoteConfig.NamespaceName, remoteConfig.NotificationID)
	}
}

// UpdateNotify update namespace's notification ID
func (n *notificationsMap) UpdateNotify(namespaceName string, notificationID int64) {
	if namespaceName != "" {
		n.setNotify(namespaceName, notificationID)
	}
}

func (n *notificationsMap) setNotify(namespaceName string, notificationID int64) {
	n.notifications.Store(namespaceName, notificationID)
}

func (n *notificationsMap) GetNotify(namespace string) int64 {
	value, ok := n.notifications.Load(namespace)
	if !ok || value == nil {
		return 0
	}
	return value.(int64)
}

func (n *notificationsMap) GetNotifyLen() int {
	s := n.notifications
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}

func (n *notificationsMap) GetNotifications() sync.Map {
	return n.notifications
}

// GetNotifies returns a JSON string of notifications for the specified namespace
// If namespace is empty, returns notifications for all namespaces
// Parameters:
//   - namespace: Target namespace, or empty string for all namespaces
//
// Returns:
//   - string: JSON representation of notifications
func (n *notificationsMap) GetNotifies(namespace string) string {
	notificationArr := make([]*Notification, 0)
	if namespace == "" {
		n.notifications.Range(func(key, value interface{}) bool {
			namespaceName := key.(string)
			notificationID := value.(int64)
			notificationArr = append(notificationArr,
				&Notification{
					NamespaceName:  namespaceName,
					NotificationID: notificationID,
				})
			return true
		})
	} else {
		notify, _ := n.notifications.LoadOrStore(namespace, defaultNotificationID)

		notificationArr = append(notificationArr,
			&Notification{
				NamespaceName:  namespace,
				NotificationID: notify.(int64),
			})
	}

	j, err := json.Marshal(notificationArr)

	if err != nil {
		return ""
	}

	return string(j)
}
