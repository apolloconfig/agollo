// Copyright 2026 Apollo Authors
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
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ConfigFileFormat identifies the representation returned by Apollo for a
// namespace. It intentionally follows Apollo Java's public format names.
type ConfigFileFormat string

const (
	ConfigFileFormatProperties ConfigFileFormat = "properties"
	ConfigFileFormatXML        ConfigFileFormat = "xml"
	ConfigFileFormatJSON       ConfigFileFormat = "json"
	ConfigFileFormatYML        ConfigFileFormat = "yml"
	ConfigFileFormatYAML       ConfigFileFormat = "yaml"
	ConfigFileFormatTXT        ConfigFileFormat = "txt"
)

// ParseConfigFileFormat converts an Apollo namespace suffix or format name to
// a format. An unknown suffix is treated as properties, as Apollo does.
func ParseConfigFileFormat(namespace string) ConfigFileFormat {
	value := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(namespace)), ".")
	switch value {
	case "properties":
		return ConfigFileFormatProperties
	case "xml":
		return ConfigFileFormatXML
	case "json":
		return ConfigFileFormatJSON
	case "yml":
		return ConfigFileFormatYML
	case "yaml":
		return ConfigFileFormatYAML
	case "txt":
		return ConfigFileFormatTXT
	}

	if index := strings.LastIndex(value, "."); index >= 0 {
		return ParseConfigFileFormat(value[index+1:])
	}
	return ConfigFileFormatProperties
}

// IsPropertiesCompatible reports whether a file format can be exposed as a
// property map without losing its primary client semantics.
func (f ConfigFileFormat) IsPropertiesCompatible() bool {
	return f == ConfigFileFormatProperties || f == ConfigFileFormatYAML || f == ConfigFileFormatYML
}

// ConfigSourceType shows where the current snapshot was loaded from.
type ConfigSourceType string

const (
	ConfigSourceRemote    ConfigSourceType = "remote"
	ConfigSourceLocalFile ConfigSourceType = "local_file"
	ConfigSourceConfigMap ConfigSourceType = "configmap"
	ConfigSourceNone      ConfigSourceType = "none"
)

// ConfigKey is the complete identity of one Apollo configuration snapshot.
// AppID is deliberately part of the identity so one process can use multiple
// Apollo applications without state or secret cross-talk.
type ConfigKey struct {
	AppID     string
	Cluster   string
	Namespace string
	Format    ConfigFileFormat
}

func (k ConfigKey) String() string {
	return fmt.Sprintf("%s/%s/%s.%s", k.AppID, k.Cluster, k.Namespace, k.Format)
}

// ConfigSnapshot is immutable once published. Values returns a defensive
// copy, so users cannot mutate a live client snapshot.
type ConfigSnapshot struct {
	Key            ConfigKey
	Values         map[string]interface{}
	Content        string
	ReleaseKey     string
	NotificationID int64
	Source         ConfigSourceType
	UpdatedAt      time.Time
}

// Clone returns a snapshot with independent metadata and value map.
func (s ConfigSnapshot) Clone() ConfigSnapshot {
	clone := s
	clone.Values = cloneValues(s.Values)
	return clone
}

func cloneValues(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{}
	}
	clone := make(map[string]interface{}, len(values))
	for key, value := range values {
		clone[key] = cloneValue(value)
	}
	return clone
}

func cloneValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []int:
		return append([]int(nil), typed...)
	case []interface{}:
		clone := make([]interface{}, len(typed))
		for i := range typed {
			clone[i] = cloneValue(typed[i])
		}
		return clone
	case map[string]interface{}:
		return cloneValues(typed)
	default:
		return value
	}
}

// Config is the thread-safe, live view of one namespace.
type Config interface {
	Key() ConfigKey
	Snapshot() ConfigSnapshot
	Lookup(key string) (string, bool)
	GetString(key, defaultValue string) string
	GetInt(key string, defaultValue int) int
	GetInt64(key string, defaultValue int64) int64
	GetFloat64(key string, defaultValue float64) float64
	GetBool(key string, defaultValue bool) bool
	GetDuration(key string, defaultValue time.Duration) time.Duration
	Keys() []string
	Source() ConfigSourceType
	Subscribe(listener ConfigChangeListenerV2, options ...SubscribeOption) (cancel func())
}

// ConfigFile is the live raw-content view of one namespace.
type ConfigFile interface {
	Key() ConfigKey
	Content() string
	HasContent() bool
	Format() ConfigFileFormat
	Source() ConfigSourceType
	AsMap() (map[string]interface{}, bool)
	Subscribe(listener ConfigFileChangeListener) (cancel func())
}

// ChangeType describes one property mutation.
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

// PropertyChange is one key-level difference between two snapshots.
type PropertyChange struct {
	OldValue interface{}
	NewValue interface{}
	Type     ChangeType
}

// ConfigChangeEvent is emitted after an atomic snapshot publication.
type ConfigChangeEvent struct {
	Key        ConfigKey
	Changes    map[string]PropertyChange
	ReleaseKey string
	Source     ConfigSourceType
	OccurredAt time.Time
}

// ConfigChangeListenerV2 receives filtered namespace change events.
type ConfigChangeListenerV2 func(ConfigChangeEvent)

// ConfigFileChangeEvent is emitted when raw namespace content changes.
type ConfigFileChangeEvent struct {
	Key        ConfigKey
	OldContent string
	NewContent string
	ReleaseKey string
	Source     ConfigSourceType
	OccurredAt time.Time
}

// ConfigFileChangeListener receives raw content changes.
type ConfigFileChangeListener func(ConfigFileChangeEvent)

// ConfigMapStore is the optional persistence adapter for Kubernetes ConfigMap
// fallback. The core deliberately depends only on this small contract; a
// Kubernetes client-go adapter can live in a separate module.
type ConfigMapStore interface {
	Load(ctx context.Context, key ConfigKey) (ConfigSnapshot, error)
	Save(ctx context.Context, snapshot ConfigSnapshot) error
}

type subscribeOptions struct {
	keys     map[string]struct{}
	prefixes []string
}

// SubscribeOption limits which changes a listener observes.
type SubscribeOption func(*subscribeOptions)

// WithInterestedKeys limits a listener to exact property names.
func WithInterestedKeys(keys ...string) SubscribeOption {
	return func(options *subscribeOptions) {
		if options.keys == nil {
			options.keys = make(map[string]struct{}, len(keys))
		}
		for _, key := range keys {
			if key != "" {
				options.keys[key] = struct{}{}
			}
		}
	}
}

// WithInterestedKeyPrefixes limits a listener to property name prefixes.
func WithInterestedKeyPrefixes(prefixes ...string) SubscribeOption {
	return func(options *subscribeOptions) {
		for _, prefix := range prefixes {
			if prefix != "" {
				options.prefixes = append(options.prefixes, prefix)
			}
		}
	}
}

func (options subscribeOptions) matches(changes map[string]PropertyChange) bool {
	if len(options.keys) == 0 && len(options.prefixes) == 0 {
		return true
	}
	for key := range changes {
		if _, ok := options.keys[key]; ok {
			return true
		}
		for _, prefix := range options.prefixes {
			if strings.HasPrefix(key, prefix) {
				return true
			}
		}
	}
	return false
}

// ConfigClient is the new instance-scoped Apollo client. Its methods are safe
// for concurrent use and it owns every goroutine it starts.
type ConfigClient interface {
	GetConfig(ctx context.Context, namespace string) (Config, error)
	GetConfigFor(ctx context.Context, appID, namespace string) (Config, error)
	GetConfigFile(ctx context.Context, namespace string, format ConfigFileFormat) (ConfigFile, error)
	GetConfigFileFor(ctx context.Context, appID, namespace string, format ConfigFileFormat) (ConfigFile, error)
	Monitor() ClientMonitor
	Close() error
}

// ClientMonitor exposes a stable, dependency-free view of client health.
type ClientMonitor interface {
	Snapshot() MonitorSnapshot
}

// MonitorSnapshot is deliberately small; exporters can map it to any metrics
// backend without forcing one into agollo's core dependency graph.
type MonitorSnapshot struct {
	StartedAt       time.Time
	ConfigCount     int
	RemoteRequests  uint64
	RemoteFailures  uint64
	LongPolls       uint64
	LongPollFailure uint64
	ListenerDrops   uint64
}

func sortedKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
