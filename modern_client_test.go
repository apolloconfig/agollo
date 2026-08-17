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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientOptionsValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		options ClientOptions
		want    string
	}{
		{name: "missing endpoint", options: ClientOptions{AppID: "sample"}, want: "ConfigServices or MetaServer"},
		{name: "invalid config service", options: ClientOptions{ConfigServices: []string{"localhost:8080"}}, want: "invalid URL"},
		{name: "unsupported scheme", options: ClientOptions{MetaServer: "ftp://apollo.example"}, want: "http or https"},
		{name: "offline without fallback", options: ClientOptions{Offline: true}, want: "CacheDir or ConfigMapStore"},
		{name: "negative poll delay", options: ClientOptions{MetaServer: "http://apollo.example", LongPollInitialDelay: -1}, want: "LongPollInitialDelay"},
		{name: "invalid retry", options: ClientOptions{MetaServer: "http://apollo.example", RetryBackoffMin: 3 * time.Second, RetryBackoffMax: time.Second}, want: "retry backoff"},
		{name: "negative listener queue", options: ClientOptions{MetaServer: "http://apollo.example", ListenerQueueSize: -1}, want: "ListenerQueueSize"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), test.options)
			if client != nil || err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewClient() = %#v, %v; want error containing %q", client, err, test.want)
			}
		})
	}
}

func TestApolloClientLoadsConfigAndAppliesProtocolParameters(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/configs/sample/default/application" {
			t.Errorf("unexpected request path: %s", request.URL.Path)
			http.Error(writer, "unexpected request path", http.StatusNotFound)
			return
		}
		query := request.URL.Query()
		if query.Get("ip") != "10.0.0.8" || query.Get("dataCenter") != "sh-az1" || query.Get("label") != "canary" {
			t.Errorf("unexpected Apollo query: %v", query)
			http.Error(writer, "unexpected Apollo query", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(request.Header.Get("Authorization"), "Apollo sample:") || request.Header.Get("Timestamp") == "" {
			t.Errorf("Apollo access-key signature was not sent: %v", request.Header)
			http.Error(writer, "missing signature", http.StatusUnauthorized)
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{
			"port": "8080", "enabled": "true", "timeout": "250ms",
		}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, ClientOptions{
		AppID:           "sample",
		AccessKeySecret: "secret",
		ClientIP:        "10.0.0.8",
		DataCenter:      "sh-az1",
		Label:           "canary",
	})
	defer client.Close()

	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if got := config.Int("port", 0); got != 8080 {
		t.Fatalf("Int(port) = %d, want 8080", got)
	}
	if got := config.Bool("enabled", false); !got {
		t.Fatal("Bool(enabled) = false, want true")
	}
	if got := config.Duration("timeout", 0); got != 250*time.Millisecond {
		t.Fatalf("Duration(timeout) = %s, want 250ms", got)
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

func TestApolloClientV6GetterAndExtensionContracts(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Agollo-Test-Signer") != "orders" {
			t.Fatalf("custom signer header = %q", request.Header.Get("X-Agollo-Test-Signer"))
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{
			"regions": "sh, bj",
			"ports":   []interface{}{8080, "9090"},
		}})
	}))
	defer server.Close()

	var selectorCalls atomic.Int32
	client := newTestClient(t, server.URL, ClientOptions{
		AppID: "orders",
		RequestSigner: func(_ string, headers http.Header, appID, _ string) error {
			headers.Set("X-Agollo-Test-Signer", appID)
			return nil
		},
		ConfigServiceSelector: func(appID string, services []string) (string, error) {
			if appID != "orders" || len(services) != 1 || services[0] != server.URL {
				t.Fatalf("selector input = %q, %v", appID, services)
			}
			selectorCalls.Add(1)
			return services[0], nil
		},
	})
	defer client.Close()

	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if got := config.StringSlice("regions", nil); strings.Join(got, ",") != "sh,bj" {
		t.Fatalf("StringSlice(regions) = %v", got)
	}
	if got := config.IntSlice("ports", nil); len(got) != 2 || got[0] != 8080 || got[1] != 9090 {
		t.Fatalf("IntSlice(ports) = %v", got)
	}

	events := make(chan ConfigChangeEvent, 1)
	cancel := config.Subscribe(func(event ConfigChangeEvent) { events <- event }, WithInterestedKeyRegexps(regexp.MustCompile(`^ports$`)))
	defer cancel()
	state := onlyModernState(t, client)
	state.publish(ConfigSnapshot{Values: map[string]interface{}{"regions": "sh,bj", "ports": "8081,9091"}, ReleaseKey: "r2", Source: ConfigSourceRemote})
	select {
	case event := <-events:
		if _, ok := event.Changes["ports"]; !ok {
			t.Fatalf("regex event = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for regexp-filtered change")
	}
	if selectorCalls.Load() != 1 {
		t.Fatalf("selector calls = %d, want 1", selectorCalls.Load())
	}
}

func TestApolloClientLoadEagerlyLoadsNamespaces(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"ready": "true"}})
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, ClientOptions{AppID: "orders"})
	defer client.Close()
	if err := client.Load(context.Background(), "application", "feature.properties"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := client.Monitor().Snapshot().ConfigCount; got != 2 {
		t.Fatalf("ConfigCount = %d, want 2", got)
	}
}

func TestApolloClientConfigFileYAMLAndRawListener(t *testing.T) {
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

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample"})
	defer client.Close()
	file, err := client.ConfigFile(context.Background(), "application", ConfigFileFormatYAML)
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

	modern := client
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

func TestApolloClientMultiAppIDAndIncrementalSync(t *testing.T) {
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

	client := newTestClient(t, server.URL, ClientOptions{
		AppID: "orders",
		AccessKeySecrets: map[string]string{
			"orders":   "orders-secret",
			"payments": "payments-secret",
		},
	})
	defer client.Close()

	orders, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("orders GetConfig() error = %v", err)
	}
	payments, err := client.ConfigForApp(context.Background(), "payments", "application")
	if err != nil {
		t.Fatalf("payments ConfigForApp() error = %v", err)
	}
	if orders.String("db.host", "") != "orders-db" || payments.String("db.host", "") != "payments-db" {
		t.Fatalf("AppId configurations were not isolated: orders=%q payments=%q", orders.String("db.host", ""), payments.String("db.host", ""))
	}

	events := make(chan ConfigChangeEvent, 1)
	orders.Subscribe(func(event ConfigChangeEvent) { events <- event }, WithInterestedKeyPrefixes("db."))
	state := findModernState(t, client, "orders")
	if err := state.reload(context.Background(), notification{ID: 4, Messages: map[string]int64{"db.host": 9}}); err != nil {
		t.Fatalf("incremental reload() error = %v", err)
	}
	if got := orders.String("db.host", ""); got != "db-2" {
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

func TestApolloClientRecoversIncrementalConfigWithoutBaseline(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if requests.Add(1) == 1 {
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r2", ConfigSyncType: "INCREMENTAL_SYNC", ConfigurationChanges: []configurationChange{{
				Key: "key", NewValue: "value", ConfigurationChangeType: "ADDED",
			}}})
			return
		}
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "full"}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample"})
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if got := config.String("key", ""); got != "full" {
		t.Fatalf("Config() value = %q, want full snapshot", got)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("remote requests = %d, want incremental response plus full retry", got)
	}
}

func TestApolloClientRetriesFailedInitialLoad(t *testing.T) {
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

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample"})
	defer client.Close()
	if _, err := client.Config(context.Background(), "application"); err == nil {
		t.Fatal("first GetConfig() succeeded, want remote error")
	}
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("second GetConfig() error = %v, want successful retry", err)
	}
	if config.String("key", "") != "recovered" {
		t.Fatalf("recovered config = %q", config.String("key", ""))
	}
	if _, err := client.Config(context.Background(), "application"); err != nil {
		t.Fatalf("third GetConfig() returned stale error: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("remote attempts = %d, want 2", got)
	}
}

func TestApolloClientAcknowledges304NotificationWithoutChangeEvent(t *testing.T) {
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

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample"})
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	events := make(chan ConfigChangeEvent, 1)
	config.Subscribe(func(event ConfigChangeEvent) { events <- event })
	state := onlyModernState(t, client)
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

func TestApolloClientRefreshFailureKeepsLastKnownGoodSnapshot(t *testing.T) {
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

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample", CacheDir: directory})
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	state := onlyModernState(t, client)
	stale := state.Snapshot()
	stale.Values["key"] = "stale"
	stale.ReleaseKey = "old"
	stale.Source = ConfigSourceLocalFile
	if err := client.persistLocalSnapshot(stale); err != nil {
		t.Fatalf("persist stale cache: %v", err)
	}

	unavailable.Store(true)
	if err := state.reload(context.Background(), notification{ID: 2}); err == nil {
		t.Fatal("reload() succeeded by rolling back to a fallback snapshot")
	}
	if config.String("key", "") != "fresh" {
		t.Fatalf("config rolled back to %q", config.String("key", ""))
	}
	if snapshot := state.Snapshot(); snapshot.Source != ConfigSourceRemote || snapshot.ReleaseKey != "r1" {
		t.Fatalf("refresh replaced last known-good snapshot: %#v", snapshot)
	}
}

func TestApolloClientCloseRejectsNewStateAndSubscriptions(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "value"}})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, ClientOptions{AppID: "sample"})
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	state := onlyModernState(t, client)
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := client.Config(context.Background(), "other"); err == nil {
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

func TestApolloClientCloseWaitsForConfigSubscription(t *testing.T) {
	client := newTestClient(t, "http://127.0.0.1", ClientOptions{AppID: "sample"})
	state := newModernConfig(client, ConfigKey{AppID: "sample", Cluster: "default", Namespace: "application", Format: ConfigFileFormatProperties})
	client.mu.Lock()
	client.states[state.key] = state
	client.mu.Unlock()

	started := make(chan struct{})
	release := make(chan struct{})
	state.Subscribe(func(ConfigChangeEvent) {
		close(started)
		<-release
	})
	state.publish(ConfigSnapshot{Values: map[string]interface{}{"key": "value"}, Source: ConfigSourceRemote})
	<-started

	closed := make(chan error, 1)
	go func() { closed <- client.Close() }()
	select {
	case err := <-closed:
		t.Fatalf("Close() returned before listener finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-closed:
		if err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close() did not wait for listener completion")
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

func TestDecodeDiskSnapshotRejectsUnsupportedVersion(t *testing.T) {
	key := ConfigKey{AppID: "sample", Cluster: "default", Namespace: "application", Format: ConfigFileFormatProperties}
	body, err := json.Marshal(diskSnapshot{Version: modernCacheVersion + 1, AppID: key.AppID, Cluster: key.Cluster, Namespace: key.Namespace})
	if err != nil {
		t.Fatalf("marshal cache snapshot: %v", err)
	}
	if _, err := decodeDiskSnapshot(key, body); err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("decodeDiskSnapshot() error = %v, want unsupported version", err)
	}
}

func TestApolloClientSerializesMetaDiscovery(t *testing.T) {
	var discoveryRequests atomic.Int32
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/services/config" {
			http.Error(writer, "unexpected path", http.StatusNotFound)
			return
		}
		discoveryRequests.Add(1)
		writeJSON(t, writer, []map[string]string{{"homepageUrl": "http://config.example"}})
	})
	server.Start()
	defer server.Close()

	client, err := NewClient(context.Background(), ClientOptions{AppID: "sample", MetaServer: server.URL, DisableLongPolling: true})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := client.configServices(context.Background(), "sample"); err != nil {
				t.Errorf("configServices() error = %v", err)
			}
		}()
	}
	group.Wait()
	if got := discoveryRequests.Load(); got != 1 {
		t.Fatalf("Meta discovery requests = %d, want 1", got)
	}
}

func TestApolloClientFallsBackToAtomicLocalCache(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "remote"}})
	}))

	remoteClient := newTestClient(t, server.URL, ClientOptions{AppID: "sample", CacheDir: directory})
	config, err := remoteClient.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("remote GetConfig() error = %v", err)
	}
	if got := config.String("key", ""); got != "remote" {
		t.Fatalf("remote key = %q", got)
	}
	if err := remoteClient.Close(); err != nil {
		t.Fatalf("remote Close() error = %v", err)
	}
	server.Close()

	local, err := NewClient(context.Background(), ClientOptions{AppID: "sample", CacheDir: directory, Offline: true})
	if err != nil {
		t.Fatalf("NewClient(local mode) error = %v", err)
	}
	defer local.Close()
	config, err = local.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("local GetConfig() error = %v", err)
	}
	if got := config.String("key", ""); got != "remote" || config.Source() != ConfigSourceLocalFile {
		t.Fatalf("local fallback = %q from %s", got, config.Source())
	}
}

func TestApolloClientPersistsAndFallsBackToConfigMapStore(t *testing.T) {
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

	remote, err := NewClient(context.Background(), ClientOptions{
		AppID: "sample", ConfigServices: []string{server.URL}, ConfigMapStore: store, DisableLongPolling: true,
	})
	if err != nil {
		t.Fatalf("NewClient(remote) error = %v", err)
	}
	if _, err := remote.Config(context.Background(), "application"); err != nil {
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
	fallback, err := NewClient(context.Background(), ClientOptions{
		AppID: "sample", ConfigServices: []string{server.URL}, ConfigMapStore: store, DisableLongPolling: true,
	})
	if err != nil {
		t.Fatalf("NewClient(fallback) error = %v", err)
	}
	defer fallback.Close()
	config, err := fallback.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("fallback GetConfig() error = %v", err)
	}
	if got := config.String("key", ""); got != "remote" || config.Source() != ConfigSourceConfigMap {
		t.Fatalf("ConfigMap fallback = %q from %s", got, config.Source())
	}
}

func TestApolloClientOfflineLoadsConfigMapWithoutCacheDirectory(t *testing.T) {
	t.Parallel()
	store := &memoryConfigMapStore{snapshot: ConfigSnapshot{
		Key:    ConfigKey{AppID: "sample", Cluster: "default", Namespace: "application", Format: ConfigFileFormatProperties},
		Values: map[string]interface{}{"key": "offline"},
	}}
	client, err := NewClient(context.Background(), ClientOptions{
		AppID: "sample", ConfigMapStore: store, Offline: true,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if got := config.String("key", ""); got != "offline" || config.Source() != ConfigSourceConfigMap {
		t.Fatalf("offline ConfigMap fallback = %q from %s", got, config.Source())
	}
}

func TestApolloClientOfflineConfigMapDoesNotReadWorkingDirectoryCache(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	temporaryDirectory := t.TempDir()
	if err := os.Chdir(temporaryDirectory); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(workingDirectory) })
	if err := os.WriteFile("sample-application.json", []byte(`{"appId":"sample","cluster":"default","namespaceName":"application","configurations":{"key":"disk"}}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := &memoryConfigMapStore{snapshot: ConfigSnapshot{
		Key:    ConfigKey{AppID: "sample", Cluster: "default", Namespace: "application", Format: ConfigFileFormatProperties},
		Values: map[string]interface{}{"key": "configmap"},
	}}
	client, err := NewClient(context.Background(), ClientOptions{AppID: "sample", ConfigMapStore: store, Offline: true})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if got := config.String("key", ""); got != "configmap" || config.Source() != ConfigSourceConfigMap {
		t.Fatalf("offline ConfigMap value = %q from %s", got, config.Source())
	}
}

func TestApolloClientLoadHonorsOperationAndLifecycleContexts(t *testing.T) {
	t.Run("operation context cancels ConfigMap load", func(t *testing.T) {
		store := &blockingConfigMapStore{started: make(chan struct{}, 1)}
		client, err := NewClient(context.Background(), ClientOptions{AppID: "sample", ConfigMapStore: store, Offline: true})
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() {
			_, err := client.Config(ctx, "application")
			result <- err
		}()
		<-store.started
		cancel()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Config() error = %v, want context cancellation", err)
			}
		case <-time.After(time.Second):
			t.Fatal("ConfigMap load did not observe operation cancellation")
		}
	})

	t.Run("client lifecycle cancels remote load", func(t *testing.T) {
		started := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
			select {
			case started <- struct{}{}:
			default:
			}
			<-request.Context().Done()
		}))
		defer server.Close()
		clientContext, cancelClient := context.WithCancel(context.Background())
		client, err := NewClient(clientContext, ClientOptions{AppID: "sample", ConfigServices: []string{server.URL}, DisableLongPolling: true})
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
		result := make(chan error, 1)
		go func() {
			_, err := client.Config(context.Background(), "application")
			result <- err
		}()
		<-started
		cancelClient()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("Config() error = %v, want context cancellation", err)
			}
		case <-time.After(time.Second):
			t.Fatal("remote load did not observe client cancellation")
		}
	})
}

func TestApolloClientIntSliceReadsNativeIntSlice(t *testing.T) {
	store := &memoryConfigMapStore{snapshot: ConfigSnapshot{
		Key:    ConfigKey{AppID: "sample", Cluster: "default", Namespace: "application", Format: ConfigFileFormatProperties},
		Values: map[string]interface{}{"ports": []int{8080, 9090}},
	}}
	client, err := NewClient(context.Background(), ClientOptions{AppID: "sample", ConfigMapStore: store, Offline: true})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if got := config.IntSlice("ports", nil); len(got) != 2 || got[0] != 8080 || got[1] != 9090 {
		t.Fatalf("IntSlice(ports) = %v", got)
	}
}

func TestApolloClientDiscoversConfigServiceFromMetaServer(t *testing.T) {
	t.Parallel()
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/services/config":
			if request.URL.Query().Get("appId") != "sample" || request.URL.Query().Get("ip") != "10.0.0.3" {
				t.Errorf("unexpected discovery query: %v", request.URL.Query())
				http.Error(writer, "unexpected discovery query", http.StatusBadRequest)
				return
			}
			writeJSON(t, writer, []map[string]string{{"homepageUrl": server.URL}})
		case "/configs/sample/default/application":
			writeJSON(t, writer, remoteConfig{ReleaseKey: "r1", Configurations: map[string]interface{}{"key": "discovered"}})
		default:
			t.Errorf("unexpected request path: %s", request.URL.Path)
			http.Error(writer, "unexpected request path", http.StatusNotFound)
		}
	})
	server.Start()
	defer server.Close()

	client, err := NewClient(context.Background(), ClientOptions{
		AppID: "sample", MetaServer: server.URL, ClientIP: "10.0.0.3", DisableLongPolling: true,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if got := config.String("key", ""); got != "discovered" {
		t.Fatalf("discovered value = %q", got)
	}
}

func TestConfigURLRejectsUnsafeIdentifiers(t *testing.T) {
	client := newTestClient(t, "http://config.example", ClientOptions{AppID: "sample"})
	defer client.Close()
	for _, key := range []ConfigKey{
		{AppID: "", Namespace: "application"},
		{AppID: "../orders", Namespace: "application"},
		{AppID: "orders", Namespace: "../application"},
		{AppID: "orders", Namespace: "team/application"},
	} {
		key.Cluster = "default"
		key.Format = ConfigFileFormatProperties
		if _, err := client.configURL("http://config.example", key, nil, nil); err == nil {
			t.Fatalf("configURL(%+v) succeeded", key)
		}
	}
}

func TestNotificationMatchesNamespacePropertiesSuffix(t *testing.T) {
	if !notificationMatchesNamespace("application.properties", "application") {
		t.Fatal("properties notification did not match extensionless namespace")
	}
	if !notificationMatchesNamespace("application", "application.properties") {
		t.Fatal("extensionless notification did not match properties namespace")
	}
	if notificationMatchesNamespace("other.properties", "application") {
		t.Fatal("different namespaces matched")
	}
}

func TestApolloClientLongPollBuildsDataCenterAndRefreshes(t *testing.T) {
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
	wrongServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		t.Fatalf("selector was bypassed for %s", request.URL.Path)
	}))
	defer wrongServer.Close()

	var selectorCalls atomic.Int32
	client, err := NewClient(context.Background(), ClientOptions{
		AppID:              "sample",
		DataCenter:         "sh",
		ClientIP:           "127.0.0.2",
		ConfigServices:     []string{wrongServer.URL, server.URL},
		DisableLongPolling: true,
		ConfigServiceSelector: func(appID string, services []string) (string, error) {
			if appID != "sample" || len(services) != 2 {
				t.Fatalf("selector input = %q, %v", appID, services)
			}
			selectorCalls.Add(1)
			return server.URL, nil
		},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()
	config, err := client.Config(context.Background(), "application")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	poller := &appPoller{client: client, appID: "sample", wake: make(chan struct{}, 1)}
	if err := poller.poll(); err != nil {
		t.Fatalf("poll() error = %v", err)
	}
	if got := config.String("key", ""); got != "after" {
		t.Fatalf("long-poll refreshed value = %q", got)
	}
	if got := selectorCalls.Load(); got != 3 {
		t.Fatalf("selector calls = %d, want initial fetch, long-poll, and refresh selection", got)
	}
}

func TestApolloClientCloseCancelsLongPoll(t *testing.T) {
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

	client, err := NewClient(context.Background(), ClientOptions{AppID: "sample", ConfigServices: []string{server.URL}})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.Config(context.Background(), "application"); err != nil {
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

func newTestClient(t *testing.T, configServiceURL string, options ClientOptions) *ApolloClient {
	t.Helper()
	options.ConfigServices = []string{configServiceURL}
	options.DisableLongPolling = true
	client, err := NewClient(context.Background(), options)
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

func onlyModernState(t *testing.T, client *ApolloClient) *modernConfig {
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

func findModernState(t *testing.T, client *ApolloClient, appID string) *modernConfig {
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

type blockingConfigMapStore struct {
	started chan struct{}
}

func (s *blockingConfigMapStore) Load(ctx context.Context, _ ConfigKey) (ConfigSnapshot, error) {
	select {
	case s.started <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return ConfigSnapshot{}, ctx.Err()
}

func (s *blockingConfigMapStore) Save(context.Context, ConfigSnapshot) error { return nil }

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
