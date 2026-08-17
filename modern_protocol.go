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
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxApolloResponseBytes = 8 << 20

var errNotModified = errors.New("agollo: configuration not modified")

type httpStatusError struct {
	StatusCode int
	URL        string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("agollo: request to %s failed with HTTP %d", e.URL, e.StatusCode)
}

type serviceSet struct {
	mu      sync.Mutex
	urls    []string
	next    uint64
	updated time.Time
	expires time.Time
}

func (set *serviceSet) choose() string {
	set.mu.Lock()
	defer set.mu.Unlock()
	if len(set.urls) == 0 {
		return ""
	}
	url := set.urls[set.next%uint64(len(set.urls))]
	set.next++
	return url
}

func (set *serviceSet) set(urls []string, ttl time.Duration) {
	set.mu.Lock()
	defer set.mu.Unlock()
	set.urls = append(set.urls[:0], urls...)
	set.updated = time.Now()
	set.expires = set.updated.Add(ttl)
}

func (set *serviceSet) stale() bool {
	set.mu.Lock()
	defer set.mu.Unlock()
	return len(set.urls) == 0 || time.Now().After(set.expires)
}

func (c *ApolloClient) loadConfig(ctx context.Context, state *modernConfig) error {
	return c.loadConfigWithNotification(ctx, state, notification{})
}

func (c *ApolloClient) loadConfigWithNotification(ctx context.Context, state *modernConfig, notification notification) error {
	loadContext, cancel := c.loadContext(ctx)
	defer cancel()

	if c.options.localMode {
		if c.options.localCacheDir != "" {
			snapshot, err := c.loadLocalSnapshot(loadContext, state.key)
			if err == nil {
				state.publish(snapshot)
				return nil
			}
		}
		return c.loadConfigMapSnapshot(loadContext, state)
	}

	requestContext, requestCancel := context.WithTimeout(loadContext, 10*time.Second)
	defer requestCancel()
	snapshot, err := c.fetchRemoteSnapshot(requestContext, state, notification)
	if err == nil {
		state.publish(snapshot)
		if c.options.localCacheDir != "" {
			if persistErr := c.persistLocalSnapshot(snapshot); persistErr != nil {
				// A cache write failure must not discard a successfully published remote snapshot.
				c.monitor.failures.Add(1)
			}
		}
		if c.options.configMapStore != nil {
			copyForStore := snapshot.Clone()
			c.goBackground(func(background context.Context) {
				if err := c.options.configMapStore.Save(background, copyForStore); err != nil {
					c.monitor.failures.Add(1)
				}
			})
		}
		return nil
	}
	if errors.Is(err, errNotModified) {
		state.acknowledgeNotification(notification.ID)
		return nil
	}
	c.monitor.failures.Add(1)
	// A fallback source establishes the first usable snapshot only. During a
	// refresh, a transient remote failure must leave the last known-good memory
	// snapshot untouched rather than rolling back to an older disk/ConfigMap
	// copy and emitting a false change event.
	if state.current().Source != ConfigSourceNone {
		return err
	}
	if c.options.localCacheDir != "" {
		if localSnapshot, localErr := c.loadLocalSnapshot(loadContext, state.key); localErr == nil {
			state.publish(localSnapshot)
			return nil
		}
	}
	if err := c.loadConfigMapSnapshot(loadContext, state); err == nil {
		return nil
	}
	return err
}

func (c *ApolloClient) loadConfigMapSnapshot(ctx context.Context, state *modernConfig) error {
	if c.options.configMapStore == nil {
		return errors.New("agollo: ConfigMap fallback is not configured")
	}
	snapshot, err := c.options.configMapStore.Load(ctx, state.key)
	if err != nil {
		return fmt.Errorf("agollo: load ConfigMap fallback: %w", err)
	}
	snapshot.Key = state.key
	snapshot.Values = cloneValues(snapshot.Values)
	snapshot.Source = ConfigSourceConfigMap
	if snapshot.Content == "" {
		snapshot.Content = extractContent(snapshot.Values, state.key.Format)
	}
	if snapshot.UpdatedAt.IsZero() {
		snapshot.UpdatedAt = time.Now()
	}
	state.publish(snapshot)
	return nil
}

func (c *ApolloClient) fetchRemoteSnapshot(ctx context.Context, state *modernConfig, notice notification) (ConfigSnapshot, error) {
	services, err := c.configServices(ctx, state.key.AppID)
	if err != nil {
		return ConfigSnapshot{}, err
	}
	previous := state.current()
	var lastErr error
	for range services {
		serviceURL, selectErr := c.selectConfigService(state.key.AppID, services)
		if selectErr != nil {
			return ConfigSnapshot{}, selectErr
		}
		if serviceURL == "" {
			break
		}
		endpoint, err := c.configURL(serviceURL, state.key, previous, notice.Messages)
		if err != nil {
			return ConfigSnapshot{}, err
		}
		result, err := c.getJSON(ctx, endpoint, state.key.AppID, 10*time.Second, nil)
		if err != nil {
			if errors.Is(err, errNotModified) {
				return ConfigSnapshot{}, errNotModified
			}
			lastErr = err
			continue
		}
		return snapshotFromRemote(state.key, previous, notice.ID, result)
	}
	if lastErr == nil {
		lastErr = errors.New("agollo: no available Config Service")
	}
	return ConfigSnapshot{}, lastErr
}

func (c *ApolloClient) selectConfigService(appID string, services []string) (string, error) {
	if c.options.serviceSelector != nil {
		selected, err := c.options.serviceSelector(appID, append([]string(nil), services...))
		if err != nil {
			return "", fmt.Errorf("agollo: select Config Service: %w", err)
		}
		selected = strings.TrimRight(strings.TrimSpace(selected), "/")
		for _, service := range services {
			if selected == service {
				return selected, nil
			}
		}
		return "", fmt.Errorf("agollo: ConfigServiceSelector returned unknown service %q", selected)
	}
	return c.nextConfigService(appID), nil
}

func (c *ApolloClient) configServices(ctx context.Context, appID string) ([]string, error) {
	if len(c.options.configServiceURLs) > 0 {
		return append([]string(nil), c.options.configServiceURLs...), nil
	}

	c.mu.Lock()
	set := c.serviceState[appID]
	if set == nil {
		set = &serviceSet{}
		c.serviceState[appID] = set
	}
	c.mu.Unlock()
	if !set.stale() {
		set.mu.Lock()
		urls := append([]string(nil), set.urls...)
		set.mu.Unlock()
		return urls, nil
	}

	metaURL := strings.TrimRight(c.options.metaServerURL, "/") + "/services/config"
	query := url.Values{}
	query.Set("appId", appID)
	if c.options.localIP != "" {
		query.Set("ip", c.options.localIP)
	}
	response, err := c.getJSON(ctx, metaURL+"?"+query.Encode(), appID, 10*time.Second, nil)
	if err != nil {
		return nil, err
	}
	var services []struct {
		HomepageURL string `json:"homepageUrl"`
	}
	if err := json.Unmarshal(response, &services); err != nil {
		return nil, fmt.Errorf("agollo: decode Config Service discovery response: %w", err)
	}
	urls := make([]string, 0, len(services))
	for _, service := range services {
		if value := strings.TrimRight(strings.TrimSpace(service.HomepageURL), "/"); value != "" {
			urls = append(urls, value)
		}
	}
	if len(urls) == 0 {
		return nil, errors.New("agollo: Meta Server returned no Config Service")
	}
	set.set(urls, 5*time.Minute)
	return urls, nil
}

func (c *ApolloClient) nextConfigService(appID string) string {
	if len(c.options.configServiceURLs) > 0 {
		c.mu.Lock()
		set := c.serviceState[appID]
		if set == nil {
			set = &serviceSet{}
			set.set(c.options.configServiceURLs, 24*time.Hour)
			c.serviceState[appID] = set
		}
		c.mu.Unlock()
		return set.choose()
	}
	c.mu.Lock()
	set := c.serviceState[appID]
	c.mu.Unlock()
	if set == nil {
		return ""
	}
	return set.choose()
}

func (c *ApolloClient) configURL(serviceURL string, key ConfigKey, previous *ConfigSnapshot, messages map[string]int64) (string, error) {
	base, err := url.Parse(serviceURL)
	if err != nil {
		return "", fmt.Errorf("agollo: invalid Config Service URL %q: %w", serviceURL, err)
	}
	base.Path = path.Join(base.Path, "configs", key.AppID, key.Cluster, key.Namespace)
	query := base.Query()
	if previous != nil && previous.ReleaseKey != "" {
		query.Set("releaseKey", previous.ReleaseKey)
	}
	if c.options.localIP != "" {
		query.Set("ip", c.options.localIP)
	}
	if c.options.dataCenter != "" {
		query.Set("dataCenter", c.options.dataCenter)
	}
	if c.options.label != "" {
		query.Set("label", c.options.label)
	}
	if len(messages) > 0 {
		encoded, err := json.Marshal(messages)
		if err != nil {
			return "", fmt.Errorf("agollo: encode remote messages: %w", err)
		}
		query.Set("messages", string(encoded))
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (c *ApolloClient) getJSON(ctx context.Context, endpoint, appID string, timeout time.Duration, headers http.Header) ([]byte, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("agollo: create request: %w", err)
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if err := c.addSignature(req, appID); err != nil {
		return nil, err
	}
	c.monitor.requests.Add(1)
	response, err := c.options.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agollo: request %s: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotModified {
		return nil, errNotModified
	}
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
		return nil, &httpStatusError{StatusCode: response.StatusCode, URL: endpoint}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxApolloResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("agollo: read response: %w", err)
	}
	if len(body) > maxApolloResponseBytes {
		return nil, fmt.Errorf("agollo: response exceeds %d bytes", maxApolloResponseBytes)
	}
	return body, nil
}

func (c *ApolloClient) addSignature(request *http.Request, appID string) error {
	secret := c.options.secret
	if configured, ok := c.options.secretsByAppID[appID]; ok {
		secret = configured
	}
	if c.options.requestSigner != nil {
		if err := c.options.requestSigner(request.URL.String(), request.Header, appID, secret); err != nil {
			return fmt.Errorf("agollo: sign request: %w", err)
		}
		return nil
	}
	if secret == "" {
		return nil
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	pathWithQuery := request.URL.EscapedPath()
	if request.URL.RawQuery != "" {
		pathWithQuery += "?" + request.URL.RawQuery
	}
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + pathWithQuery))
	request.Header.Set("Authorization", "Apollo "+appID+":"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	request.Header.Set("Timestamp", timestamp)
	return nil
}

type remoteConfig struct {
	AppID                string                 `json:"appId"`
	Cluster              string                 `json:"cluster"`
	NamespaceName        string                 `json:"namespaceName"`
	ReleaseKey           string                 `json:"releaseKey"`
	Configurations       map[string]interface{} `json:"configurations"`
	ConfigSyncType       string                 `json:"configSyncType"`
	ConfigurationChanges []configurationChange  `json:"configurationChanges"`
}

type configurationChange struct {
	Key                     string      `json:"key"`
	NewValue                interface{} `json:"newValue"`
	ConfigurationChangeType string      `json:"configurationChangeType"`
}

func snapshotFromRemote(key ConfigKey, previous *ConfigSnapshot, notificationID int64, body []byte) (ConfigSnapshot, error) {
	var response remoteConfig
	if err := json.Unmarshal(body, &response); err != nil {
		return ConfigSnapshot{}, fmt.Errorf("agollo: decode config response: %w", err)
	}
	values := cloneValues(response.Configurations)
	syncType := strings.ToUpper(strings.TrimSpace(response.ConfigSyncType))
	if syncType == "INCREMENTAL_SYNC" {
		if previous == nil || previous.Source == ConfigSourceNone || previous.ReleaseKey == "" {
			return ConfigSnapshot{}, errors.New("agollo: received incremental config without a full snapshot baseline")
		}
		merged, err := mergeIncremental(previous.Values, response.ConfigurationChanges)
		if err != nil {
			return ConfigSnapshot{}, err
		}
		values = merged
	} else if syncType != "" && syncType != "FULL_SYNC" {
		return ConfigSnapshot{}, fmt.Errorf("agollo: unsupported config sync type %q", response.ConfigSyncType)
	}

	content := extractContent(values, key.Format)
	if key.Format.IsPropertiesCompatible() && content != "" && len(values) == 1 {
		// YAML/YML ConfigFile namespaces are represented by a single content key.
		// The raw content remains authoritative even when it cannot be flattened.
		if parsed, err := parsePropertiesCompatibleContent(key.Format, content); err == nil && len(parsed) > 0 {
			values = parsed
		}
	}
	return ConfigSnapshot{
		Key:            key,
		Values:         values,
		Content:        content,
		ReleaseKey:     response.ReleaseKey,
		NotificationID: notificationID,
		Source:         ConfigSourceRemote,
		UpdatedAt:      time.Now(),
	}, nil
}

func mergeIncremental(previous map[string]interface{}, changes []configurationChange) (map[string]interface{}, error) {
	values := cloneValues(previous)
	for _, change := range changes {
		switch strings.ToUpper(strings.TrimSpace(change.ConfigurationChangeType)) {
		case "ADDED", "MODIFIED":
			if change.Key == "" {
				return nil, errors.New("agollo: incremental configuration change has an empty key")
			}
			values[change.Key] = cloneValue(change.NewValue)
		case "DELETED":
			if change.Key == "" {
				return nil, errors.New("agollo: incremental configuration change has an empty key")
			}
			delete(values, change.Key)
		default:
			return nil, fmt.Errorf("agollo: unsupported incremental change type %q", change.ConfigurationChangeType)
		}
	}
	return values, nil
}

func extractContent(values map[string]interface{}, format ConfigFileFormat) string {
	if raw, ok := values["content"]; ok {
		if content, ok := stringify(raw); ok {
			return content
		}
	}
	if format == ConfigFileFormatProperties {
		return renderProperties(values)
	}
	return ""
}

func renderProperties(values map[string]interface{}) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "content" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		value, ok := stringify(values[key])
		if !ok {
			continue
		}
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(value)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func parsePropertiesCompatibleContent(format ConfigFileFormat, content string) (map[string]interface{}, error) {
	if format == ConfigFileFormatProperties {
		return parseProperties(content), nil
	}
	// Keep YAML parsing behind Viper, which is already an agollo dependency.
	// A new Viper instance per call avoids the package-global parser race in v5.
	return parseYAML(content, string(format))
}

func parseProperties(content string) map[string]interface{} {
	values := make(map[string]interface{})
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		index := strings.IndexAny(line, "=:")
		if index <= 0 {
			continue
		}
		values[strings.TrimSpace(line[:index])] = strings.TrimSpace(line[index+1:])
	}
	return values
}

func retryDelay(min, max time.Duration, attempt int) time.Duration {
	if attempt <= 0 {
		return min
	}
	delay := min
	for index := 0; index < attempt && delay < max; index++ {
		delay *= 2
		if delay > max {
			delay = max
		}
	}
	if delay <= 1 {
		return delay
	}
	return delay/2 + time.Duration(rand.Int63n(int64(delay/2)))
}
