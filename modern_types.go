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
	"net/http"
	"regexp"
	"sort"
	"strconv"
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
	String(key, defaultValue string) string
	Int(key string, defaultValue int) int
	Int64(key string, defaultValue int64) int64
	Float64(key string, defaultValue float64) float64
	Bool(key string, defaultValue bool) bool
	Duration(key string, defaultValue time.Duration) time.Duration
	StringSlice(key string, defaultValue []string) []string
	IntSlice(key string, defaultValue []int) []int
	Keys() []string
	Source() ConfigSourceType
	Subscribe(listener ConfigChangeHandler, options ...SubscribeOption) (cancel func())
}

// ConfigFile is the live raw-content view of one namespace.
type ConfigFile interface {
	Key() ConfigKey
	Content() string
	HasContent() bool
	Format() ConfigFileFormat
	Source() ConfigSourceType
	AsMap() (map[string]interface{}, bool)
	Subscribe(listener ConfigFileChangeHandler) (cancel func())
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

// ConfigChangeHandler receives filtered namespace change events.
type ConfigChangeHandler func(ConfigChangeEvent)

// ConfigFileChangeEvent is emitted when raw namespace content changes.
type ConfigFileChangeEvent struct {
	Key        ConfigKey
	OldContent string
	NewContent string
	ReleaseKey string
	Source     ConfigSourceType
	OccurredAt time.Time
}

// ConfigFileChangeHandler receives raw content changes.
type ConfigFileChangeHandler func(ConfigFileChangeEvent)

// ConfigMapStore is the optional persistence adapter for Kubernetes ConfigMap
// fallback. The core deliberately depends only on this small contract; a
// Kubernetes client-go adapter can live in a separate module.
type ConfigMapStore interface {
	Load(ctx context.Context, key ConfigKey) (ConfigSnapshot, error)
	Save(ctx context.Context, snapshot ConfigSnapshot) error
}

// RequestSigner adds authentication headers to an outgoing Apollo request.
// It is invoked after the request has been built and before it is sent. When
// nil, Apollo's standard Access Key signature is used when a secret is set.
// Implementations must be safe for concurrent use.
type RequestSigner func(requestURL string, headers http.Header, appID, secret string) error

// ConfigServiceSelector selects one Config Service from the currently known
// service set. It replaces the legacy process-global load balancer with an
// instance-scoped policy. A nil selector uses round-robin selection.
type ConfigServiceSelector func(appID string, services []string) (string, error)

type subscribeOptions struct {
	keys     map[string]struct{}
	prefixes []string
	patterns []*regexp.Regexp
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

// WithInterestedKeyRegexps limits a listener to changes whose keys match at
// least one compiled regular expression. Supplying compiled expressions keeps
// Subscribe free of delayed validation failures.
func WithInterestedKeyRegexps(patterns ...*regexp.Regexp) SubscribeOption {
	return func(options *subscribeOptions) {
		for _, pattern := range patterns {
			if pattern != nil {
				options.patterns = append(options.patterns, pattern)
			}
		}
	}
}

func (options subscribeOptions) matches(changes map[string]PropertyChange) bool {
	if len(options.keys) == 0 && len(options.prefixes) == 0 && len(options.patterns) == 0 {
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
		for _, pattern := range options.patterns {
			if pattern.MatchString(key) {
				return true
			}
		}
	}
	return false
}

func stringSlice(value interface{}) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...), true
	case []interface{}:
		result := make([]string, len(typed))
		for index, item := range typed {
			text, ok := stringify(item)
			if !ok {
				return nil, false
			}
			result[index] = text
		}
		return result, true
	case string:
		if typed == "" {
			return []string{}, true
		}
		parts := strings.Split(typed, ",")
		for index := range parts {
			parts[index] = strings.TrimSpace(parts[index])
		}
		return parts, true
	default:
		return nil, false
	}
}

func intSlice(value interface{}) ([]int, bool) {
	strings, ok := stringSlice(value)
	if !ok {
		return nil, false
	}
	result := make([]int, len(strings))
	for index, value := range strings {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, false
		}
		result[index] = parsed
	}
	return result, true
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
