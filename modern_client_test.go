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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestModernClientLoadsConfigAndAppliesProtocolParameters(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/configs/sample/default/application" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("ip") != "10.0.0.8" || query.Get("dataCenter") != "sh-az1" || query.Get("label") != "canary" {
			t.Fatalf("unexpected Apollo query: %v", query)
		}
		if !strings.HasPrefix(request.Header.Get("Authorization"), "Apollo sample:") || request.Header.Get("Timestamp") == "" {
			t.Fatalf("Apollo access-key signature was not sent: %v", request.Header)
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{
			"port": "8080", "enabled": "true", "timeout": "250ms",
		}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL,
		WithAppID("sample"),
		WithAccessKeySecret("secret"),
		WithLocalIP("10.0.0.8"),
		WithDataCenter("sh-az1"),
		WithLabel("canary"),
	)
	defer client.Close()

	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if got := config.GetInt("port", 0); got != 8080 {
		t.Fatalf("GetInt(port) = %d, want 8080", got)
	}
	if got := config.GetBool("enabled", false); !got {
		t.Fatal("GetBool(enabled) = false, want true")
	}
	if got := config.GetDuration("timeout", 0); got != 250*time.Millisecond {
		t.Fatalf("GetDuration(timeout) = %s, want 250ms", got)
	}
	if got := config.Source(); got != ConfigSourceRemote {
		t.Fatalf("Source() = %s, want remote", got)
	}
	if got := config.Keys(); strings.Join(got, ",") != "enabled,port,timeout" {
		t.Fatalf("Keys() = %v", got)
	}
	if monitor := client.Monitor().Snapshot(); monitor.RemoteRequests != 1 || monitor.ConfigCount != 1 {
		t.Fatalf("unexpected monitor snapshot: %+v", monitor)
	}
}

func TestModernClientConfigFileYAMLAndRawListener(t *testing.T) {
	t.Parallel()
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/configs/sample/default/application.yaml" {
			t.Fatalf("unexpected config file path: %s", request.URL.Path)
		}
		requestCount++
		content := "server:\n  port: 8080\n"
		if requestCount > 1 {
			content = "server:\n  port: 9090\n"
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r" + string(rune('0'+requestCount)), Configurations: map[string]interface{}{"content": content}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"))
	defer client.Close()
	file, err := client.GetConfigFile(context.Background(), "application", ConfigFileFormatYAML)
	if err != nil {
		t.Fatalf("GetConfigFile() error = %v", err)
	}
	if !file.HasContent() || !strings.Contains(file.Content(), "8080") {
		t.Fatalf("unexpected ConfigFile content: %q", file.Content())
	}
	properties, ok := file.AsMap()
	if !ok || properties["server.port"] != 8080 {
		t.Fatalf("AsMap() = %#v, %v", properties, ok)
	}
	changes := make(chan ConfigFileChangeEvent, 1)
	cancel := file.Subscribe(func(event ConfigFileChangeEvent) { changes <- event })
	defer cancel()

	modern := client.(*modernClient)
	state := onlyModernState(t, modern)
	if err := state.reload(context.Background(), notification{ID: 2}); err != nil {
		t.Fatalf("reload() error = %v", err)
	}
	select {
	case event := <-changes:
		if !strings.Contains(event.OldContent, "8080") || !strings.Contains(event.NewContent, "9090") {
			t.Fatalf("unexpected raw content event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ConfigFile change event")
	}
}

func TestModernClientMultiAppIDAndIncrementalSync(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path, "/")
		if len(parts) != 5 || parts[1] != "configs" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		appID := parts[2]
		mu.Lock()
		requests[appID]++
		count := requests[appID]
		mu.Unlock()
		if !strings.HasPrefix(request.Header.Get("Authorization"), "Apollo "+appID+":") {
			t.Fatalf("request for %s did not use AppId-specific signature", appID)
		}
		if appID == "orders" && count > 1 {
			if request.URL.Query().Get("releaseKey") != "orders-r1" {
				t.Fatalf("incremental request releaseKey = %q", request.URL.Query().Get("releaseKey"))
			}
			messages := request.URL.Query().Get("messages")
			if messages != "{\"db.host\":9}" {
				t.Fatalf("incremental request messages = %q", messages)
			}
			writeJSON(t, writer, remoteConfig{ReleaseKey: "orders-r2", ConfigSyncType: "INCREMENTAL_SYNC", ConfigurationChanges: []configurationChange{
				{Key: "db.host", NewValue: "db-2", ConfigurationChangeType: "MODIFIED"},
				{Key: "removed", ConfigurationChangeType: "DELETED"},
			}})
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: appID + "-r1", Configurations: map[string]interface{}{
			"db.host": appID + "-db", "removed": "yes",
		}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL,
		WithAppID("orders"),
		WithAppIDAccessKeySecret("orders", "orders-secret"),
		WithAppIDAccessKeySecret("payments", "payments-secret"),
	)
	defer client.Close()

	orders, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("orders GetConfig() error = %v", err)
	}
	payments, err := client.GetConfigFor(context.Background(), "payments", "application")
	if err != nil {
		t.Fatalf("payments GetConfigFor() error = %v", err)
	}
	if orders.GetString("db.host", "") != "orders-db" || payments.GetString("db.host", "") != "payments-db" {
		t.Fatalf("AppId configurations were not isolated: orders=%q payments=%q", orders.GetString("db.host", ""), payments.GetString("db.host", ""))
	}

	events := make(chan ConfigChangeEvent, 1)
	orders.Subscribe(func(event ConfigChangeEvent) { events <- event }, WithInterestedKeyPrefixes("db."))
	state := findModernState(t, client.(*modernClient), "orders")
	if err := state.reload(context.Background(), notification{ID: 4, Messages: map[string]int64{"db.host": 9}}); err != nil {
		t.Fatalf("incremental reload() error = %v", err)
	}
	if got := orders.GetString("db.host", ""); got != "db-2" {
		t.Fatalf("incremental value = %q, want db-2", got)
	}
	if _, exists := orders.Lookup("removed"); exists {
		t.Fatal("deleted incremental key still exists")
	}
	select {
	case event := <-events:
		if len(event.Changes) != 2 || event.Changes["db.host"].Type != ChangeModified {
			t.Fatalf("unexpected filtered change event: %#v", event)
		}
		if _, exists := event.Changes["removed"]; !exists {
			t.Fatal("filtered event must retain the complete atomic change set")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for incremental event")
	}
}

func TestModernClientRejectsIncrementalConfigWithoutBaseline(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r2", ConfigSyncType: "INCREMENTAL_SYNC", ConfigurationChanges: []configurationChange{{
			Key: "key", NewValue: "value", ConfigurationChangeType: "ADDED",
		}}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"))
	defer client.Close()
	if _, err := client.GetConfig(context.Background(), "application"); err == nil || !strings.Contains(err.Error(), "without a full snapshot baseline") {
		t.Fatalf("GetConfig() error = %v, want incremental baseline error", err)
	}
}

func TestModernClientRetriesFailedInitialLoad(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if attempts.Add(1) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "recovered"}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"))
	defer client.Close()
	if _, err := client.GetConfig(context.Background(), "application"); err == nil {
		t.Fatal("first GetConfig() succeeded, want remote error")
	}
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("second GetConfig() error = %v, want successful retry", err)
	}
	if config.GetString("key", "") != "recovered" {
		t.Fatalf("recovered config = %q", config.GetString("key", ""))
	}
	if _, err := client.GetConfig(context.Background(), "application"); err != nil {
		t.Fatalf("third GetConfig() returned stale error: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("remote attempts = %d, want 2", got)
	}
}

func TestModernClientAcknowledges304NotificationWithoutChangeEvent(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if requests.Add(1) == 1 {
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "stable"}})
			return
		}
		if request.URL.Query().Get("releaseKey") != "r1" {
			t.Fatalf("304 request releaseKey = %q", request.URL.Query().Get("releaseKey"))
		}
		writer.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"))
	defer client.Close()
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	events := make(chan ConfigChangeEvent, 1)
	config.Subscribe(func(event ConfigChangeEvent) { events <- event })
	state := onlyModernState(t, client.(*modernClient))
	if err := state.reload(context.Background(), notification{ID: 9}); err != nil {
		t.Fatalf("reload() error = %v", err)
	}
	snapshot := state.Snapshot()
	if snapshot.NotificationID != 9 || snapshot.ReleaseKey != "r1" || snapshot.Values["key"] != "stable" {
		t.Fatalf("304 changed snapshot unexpectedly: %#v", snapshot)
	}
	select {
	case event := <-events:
		t.Fatalf("304 emitted change event: %#v", event)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestModernClientRefreshFailureKeepsLastKnownGoodSnapshot(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	var unavailable atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if unavailable.Load() {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "fresh"}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"), WithLocalCacheDir(directory))
	defer client.Close()
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	state := onlyModernState(t, client.(*modernClient))
	stale := state.Snapshot()
	stale.Values["key"] = "stale"
	stale.ReleaseKey = "old"
	stale.Source = ConfigSourceLocalFile
	if err := client.(*modernClient).persistLocalSnapshot(stale); err != nil {
		t.Fatalf("persist stale cache: %v", err)
	}

	unavailable.Store(true)
	if err := state.reload(context.Background(), notification{ID: 2}); err == nil {
		t.Fatal("reload() succeeded by rolling back to a fallback snapshot")
	}
	if config.GetString("key", "") != "fresh" {
		t.Fatalf("config rolled back to %q", config.GetString("key", ""))
	}
	if snapshot := state.Snapshot(); snapshot.Source != ConfigSourceRemote || snapshot.ReleaseKey != "r1" {
		t.Fatalf("refresh replaced last known-good snapshot: %#v", snapshot)
	}
}

func TestModernClientCloseRejectsNewStateAndSubscriptions(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "value"}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"))
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	state := onlyModernState(t, client.(*modernClient))
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := client.GetConfig(context.Background(), "other"); err == nil {
		t.Fatal("GetConfig() after Close succeeded")
	}
	config.Subscribe(func(ConfigChangeEvent) {})
	(*modernConfigFile)(state).Subscribe(func(ConfigFileChangeEvent) {})
	state.listenerMu.Lock()
	configListeners, fileListeners := len(state.listeners), len(state.fileListeners)
	state.listenerMu.Unlock()
	if configListeners != 0 || fileListeners != 0 {
		t.Fatalf("subscriptions registered after Close: config=%d file=%d", configListeners, fileListeners)
	}
}

func TestDecodeDiskSnapshotRejectsDifferentCluster(t *testing.T) {
	t.Parallel()
	body, err := json.Marshal(diskSnapshot{
		AppID:     "sample",
		Cluster:   "production",
		Namespace: "application",
		Values:    map[string]interface{}{"key": "value"},
	})
	if err != nil {
		t.Fatalf("marshal cache snapshot: %v", err)
	}
	_, err = decodeDiskSnapshot(ConfigKey{AppID: "sample", Cluster: "staging", Namespace: "application", Format: ConfigFileFormatProperties}, body)
	if err == nil || !strings.Contains(err.Error(), "cluster") {
		t.Fatalf("decodeDiskSnapshot() error = %v, want cluster mismatch", err)
	}
}

func TestModernClientFallsBackToAtomicLocalCache(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "remote"}})
	}))

	remoteClient := newTestClient(t, server.URL, WithAppID("sample"), WithLocalCacheDir(directory))
	config, err := remoteClient.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("remote GetConfig() error = %v", err)
	}
	if got := config.GetString("key", ""); got != "remote" {
		t.Fatalf("remote key = %q", got)
	}
	if err := remoteClient.Close(); err != nil {
		t.Fatalf("remote Close() error = %v", err)
	}
	server.Close()

	local, err := NewClient(context.Background(), WithAppID("sample"), WithLocalCacheDir(directory), WithLocalMode())
	if err != nil {
		t.Fatalf("NewClient(local mode) error = %v", err)
	}
	defer local.Close()
	config, err = local.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("local GetConfig() error = %v", err)
	}
	if got := config.GetString("key", ""); got != "remote" || config.Source() != ConfigSourceLocalFile {
		t.Fatalf("local fallback = %q from %s", got, config.Source())
	}
}

func TestModernClientPersistsAndFallsBackToConfigMapStore(t *testing.T) {
	t.Parallel()
	store := &memoryConfigMapStore{saved: make(chan ConfigSnapshot, 1)}
	var status atomic.Int32
	status.Store(http.StatusOK)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if currentStatus := int(status.Load()); currentStatus != http.StatusOK {
			writer.WriteHeader(currentStatus)
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "remote"}})
	}))
	defer server.Close()

	remote, err := NewClient(context.Background(), WithAppID("sample"), WithConfigServiceURLs(server.URL), WithConfigMapStore(store), WithoutLongPoll())
	if err != nil {
		t.Fatalf("NewClient(remote) error = %v", err)
	}
	if _, err := remote.GetConfig(context.Background(), "application"); err != nil {
		t.Fatalf("remote GetConfig() error = %v", err)
	}
	select {
	case snapshot := <-store.saved:
		if snapshot.Values["key"] != "remote" || snapshot.Source != ConfigSourceRemote {
			t.Fatalf("unexpected ConfigMap save snapshot: %#v", snapshot)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ConfigMap save")
	}
	if err := remote.Close(); err != nil {
		t.Fatalf("remote Close() error = %v", err)
	}

	status.Store(http.StatusServiceUnavailable)
	fallback, err := NewClient(context.Background(), WithAppID("sample"), WithConfigServiceURLs(server.URL), WithConfigMapStore(store), WithoutLongPoll())
	if err != nil {
		t.Fatalf("NewClient(fallback) error = %v", err)
	}
	defer fallback.Close()
	config, err := fallback.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("fallback GetConfig() error = %v", err)
	}
	if got := config.GetString("key", ""); got != "remote" || config.Source() != ConfigSourceConfigMap {
		t.Fatalf("ConfigMap fallback = %q from %s", got, config.Source())
	}
}

func TestModernClientDiscoversConfigServiceFromMetaServer(t *testing.T) {
	t.Parallel()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/services/config":
			if request.URL.Query().Get("appId") != "sample" || request.URL.Query().Get("ip") != "10.0.0.3" {
				t.Fatalf("unexpected discovery query: %v", request.URL.Query())
			}
			writeJSON(t, writer, []map[string]string{{"homepageUrl": server.URL}})
		case "/configs/sample/default/application":
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "discovered"}})
		default:
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAppID("sample"), WithMetaServerURL(server.URL), WithLocalIP("10.0.0.3"), WithoutLongPoll())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if got := config.GetString("key", ""); got != "discovered" {
		t.Fatalf("discovered value = %q", got)
	}
}

func TestModernClientLongPollBuildsDataCenterAndRefreshes(t *testing.T) {
	t.Parallel()
	var configRequests int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/configs/sample/default/application":
			configRequests++
			if configRequests == 1 {
				writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "before"}})
				return
			}
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r2", Configurations: map[string]interface{}{"key": "after"}})
		case "/notifications/v2":
			query := request.URL.Query()
			if query.Get("dataCenter") != "sh" || query.Get("ip") != "127.0.0.2" {
				t.Fatalf("unexpected long poll query: %v", query)
			}
			var notices []notification
			if err := json.Unmarshal([]byte(query.Get("notifications")), &notices); err != nil || len(notices) != 1 || notices[0].ID != 0 {
				t.Fatalf("unexpected notifications payload %q: %#v, %v", query.Get("notifications"), notices, err)
			}
			writeJSON(t, writer, []notification{{NamespaceName: "application", ID: 2}})
		default:
			t.Fatalf("unexpected long poll request: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithAppID("sample"), WithDataCenter("sh"), WithLocalIP("127.0.0.2"))
	defer client.Close()
	config, err := client.GetConfig(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	poller := &appPoller{client: client.(*modernClient), appID: "sample", wake: make(chan struct{}, 1)}
	if err := poller.poll(); err != nil {
		t.Fatalf("poll() error = %v", err)
	}
	if got := config.GetString("key", ""); got != "after" {
		t.Fatalf("long-poll refreshed value = %q", got)
	}
}

func TestModernClientCloseCancelsLongPoll(t *testing.T) {
	t.Parallel()
	pollStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/configs/sample/default/application":
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "value"}})
		case "/notifications/v2":
			select {
			case pollStarted <- struct{}{}:
			default:
			}
			<-request.Context().Done()
		default:
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAppID("sample"), WithConfigServiceURLs(server.URL))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.GetConfig(context.Background(), "application"); err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	select {
	case <-pollStarted:
	case <-time.After(time.Second):
		t.Fatal("long poll did not start")
	}
	closed := make(chan error, 1)
	go func() { closed <- client.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() did not cancel long poll promptly")
	}
}

func newTestClient(t *testing.T, configServiceURL string, options ...ClientOption) ConfigClient {
	t.Helper()
	options = append(options, WithConfigServiceURLs(configServiceURL), WithoutLongPoll())
	client, err := NewClient(context.Background(), options...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value interface{}) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatalf("encode test response: %v", err)
	}
}

func onlyModernState(t *testing.T, client *modernClient) *modernConfig {
	t.Helper()
	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.states) != 1 {
		t.Fatalf("state count = %d, want 1", len(client.states))
	}
	for _, state := range client.states {
		return state
	}
	return nil
}

func findModernState(t *testing.T, client *modernClient, appID string) *modernConfig {
	t.Helper()
	client.mu.Lock()
	defer client.mu.Unlock()
	for key, state := range client.states {
		if key.AppID == appID {
			return state
		}
	}
	t.Fatalf("no state found for appId %s", appID)
	return nil
}

type memoryConfigMapStore struct {
	mu       sync.Mutex
	snapshot ConfigSnapshot
	saved    chan ConfigSnapshot
}

func (s *memoryConfigMapStore) Load(ctx context.Context, key ConfigKey) (ConfigSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot.Key.AppID == "" {
		return ConfigSnapshot{}, context.DeadlineExceeded
	}
	return s.snapshot.Clone(), nil
}

func (s *memoryConfigMapStore) Save(ctx context.Context, snapshot ConfigSnapshot) error {
	s.mu.Lock()
	s.snapshot = snapshot.Clone()
	s.mu.Unlock()
	select {
	case s.saved <- snapshot.Clone():
	default:
	}
	return nil
}
