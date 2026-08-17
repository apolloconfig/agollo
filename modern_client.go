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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultModernCluster = "default"

// ClientOptions configures an ApolloClient. Its zero values select safe
// defaults: cluster "default", long polling enabled, a standard HTTP client,
// 1s-2m retry backoff, and a listener queue size of 32.
//
// AppID may be empty only when every lookup uses ConfigForApp or
// ConfigFileForApp. ConfigServices takes precedence over MetaServer.
type ClientOptions struct {
	// AppID is the default application identity. ConfigForApp can override it.
	AppID string
	// Cluster defaults to "default".
	Cluster string
	// ConfigServices contains direct Config Service base URLs. When non-empty,
	// it takes precedence over MetaServer discovery.
	ConfigServices []string
	// MetaServer is the base URL used for Config Service discovery.
	MetaServer string
	// AccessKeySecret signs requests that have no AppID-specific override.
	AccessKeySecret string
	// AccessKeySecrets optionally overrides the secret by AppID.
	AccessKeySecrets map[string]string
	// Label selects an Apollo gray release.
	Label string
	// DataCenter and ClientIP are sent during fetches and long polling.
	DataCenter string
	ClientIP   string
	// HTTPClient is shared by discovery, fetch, and long-poll requests.
	HTTPClient *http.Client
	// RequestSigner overrides the default Apollo Access Key signer for this
	// client instance. It is useful for custom gateway authentication.
	RequestSigner RequestSigner
	// ConfigServiceSelector selects a Config Service for each request. A nil
	// selector uses the built-in round-robin policy.
	ConfigServiceSelector ConfigServiceSelector
	// CacheDir enables durable local cache fallback.
	CacheDir string
	// ConfigMapStore enables an optional fallback after local files.
	ConfigMapStore ConfigMapStore
	// Offline disables discovery, fetch, and long-poll network traffic. It
	// requires CacheDir or ConfigMapStore.
	Offline bool
	// DisableLongPolling keeps remote reads on demand only.
	DisableLongPolling bool
	// LongPollInitialDelay delays polling after the first successful load.
	LongPollInitialDelay time.Duration
	// RetryBackoffMin and RetryBackoffMax bound exponential retry delays.
	RetryBackoffMin time.Duration
	RetryBackoffMax time.Duration
	// ListenerQueueSize defaults to 32. Zero selects the default.
	ListenerQueueSize int
}

type clientOptions struct {
	appID             string
	cluster           string
	configServiceURLs []string
	metaServerURL     string
	secret            string
	secretsByAppID    map[string]string
	label             string
	dataCenter        string
	localIP           string
	httpClient        *http.Client
	requestSigner     RequestSigner
	serviceSelector   ConfigServiceSelector
	localCacheDir     string
	configMapStore    ConfigMapStore
	localMode         bool
	longPollEnabled   bool
	pollInitialDelay  time.Duration
	retryMin          time.Duration
	retryMax          time.Duration
	listenerQueueSize int
}

func defaultClientOptions() clientOptions {
	return clientOptions{
		cluster:           defaultModernCluster,
		httpClient:        &http.Client{},
		longPollEnabled:   true,
		retryMin:          time.Second,
		retryMax:          2 * time.Minute,
		listenerQueueSize: 32,
	}
}

func resolveClientOptions(input ClientOptions) (clientOptions, error) {
	options := defaultClientOptions()
	options.appID = strings.TrimSpace(input.AppID)
	if cluster := strings.TrimSpace(input.Cluster); cluster != "" {
		options.cluster = cluster
	}
	options.configServiceURLs = normalizeURLs(input.ConfigServices)
	for _, service := range options.configServiceURLs {
		if err := validateServerURL("ConfigServices", service); err != nil {
			return clientOptions{}, err
		}
	}
	options.metaServerURL = strings.TrimRight(strings.TrimSpace(input.MetaServer), "/")
	if options.metaServerURL != "" {
		if err := validateServerURL("MetaServer", options.metaServerURL); err != nil {
			return clientOptions{}, err
		}
	}
	options.secret = input.AccessKeySecret
	options.secretsByAppID = make(map[string]string, len(input.AccessKeySecrets))
	for rawAppID, secret := range input.AccessKeySecrets {
		appID := strings.TrimSpace(rawAppID)
		if appID == "" {
			return clientOptions{}, errors.New("agollo: AccessKeySecrets contains an empty AppID")
		}
		if _, exists := options.secretsByAppID[appID]; exists {
			return clientOptions{}, fmt.Errorf("agollo: AccessKeySecrets contains duplicate AppID %q after trimming", appID)
		}
		options.secretsByAppID[appID] = secret
	}
	options.label = strings.TrimSpace(input.Label)
	options.dataCenter = strings.TrimSpace(input.DataCenter)
	options.localIP = strings.TrimSpace(input.ClientIP)
	if input.HTTPClient != nil {
		options.httpClient = input.HTTPClient
	}
	options.requestSigner = input.RequestSigner
	options.serviceSelector = input.ConfigServiceSelector
	options.localCacheDir = strings.TrimSpace(input.CacheDir)
	options.configMapStore = input.ConfigMapStore
	options.localMode = input.Offline
	options.longPollEnabled = !input.DisableLongPolling && !input.Offline
	if input.LongPollInitialDelay < 0 {
		return clientOptions{}, errors.New("agollo: LongPollInitialDelay cannot be negative")
	}
	options.pollInitialDelay = input.LongPollInitialDelay
	if input.RetryBackoffMin != 0 {
		options.retryMin = input.RetryBackoffMin
	}
	if input.RetryBackoffMax != 0 {
		options.retryMax = input.RetryBackoffMax
	}
	if options.retryMin <= 0 || options.retryMax < options.retryMin {
		return clientOptions{}, errors.New("agollo: invalid retry backoff range")
	}
	if input.ListenerQueueSize < 0 {
		return clientOptions{}, errors.New("agollo: ListenerQueueSize cannot be negative")
	}
	if input.ListenerQueueSize > 0 {
		options.listenerQueueSize = input.ListenerQueueSize
	}
	return options, nil
}

func validateServerURL(field, value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("agollo: %s contains invalid URL %q", field, value)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("agollo: %s URL %q must use http or https", field, value)
	}
	return nil
}

// NewClient creates an instance-scoped Apollo client. It does not perform
// network I/O until Config or ConfigFile is called.
func NewClient(ctx context.Context, input ClientOptions) (*ApolloClient, error) {
	if ctx == nil {
		return nil, errors.New("agollo: context is nil")
	}
	config, err := resolveClientOptions(input)
	if err != nil {
		return nil, err
	}

	if config.localMode && config.localCacheDir == "" && config.configMapStore == nil {
		return nil, errors.New("agollo: ClientOptions.Offline requires CacheDir or ConfigMapStore")
	}
	if !config.localMode && len(config.configServiceURLs) == 0 && config.metaServerURL == "" {
		return nil, errors.New("agollo: configure ClientOptions.ConfigServices or MetaServer")
	}
	if config.localCacheDir != "" {
		if err := os.MkdirAll(config.localCacheDir, 0o750); err != nil {
			return nil, fmt.Errorf("agollo: create local cache directory: %w", err)
		}
	}

	clientContext, cancel := context.WithCancel(ctx)
	client := &ApolloClient{
		options:      config,
		ctx:          clientContext,
		cancel:       cancel,
		states:       make(map[ConfigKey]*modernConfig),
		appPollers:   make(map[string]*appPoller),
		monitor:      modernMonitor{startedAt: time.Now()},
		serviceState: make(map[string]*serviceSet),
	}
	return client, nil
}

// ApolloClient is the instance-scoped API. It is safe for concurrent use and
// owns every background task it starts. Close is idempotent.
type ApolloClient struct {
	options clientOptions
	ctx     context.Context
	cancel  context.CancelFunc

	mu           sync.Mutex
	states       map[ConfigKey]*modernConfig
	appPollers   map[string]*appPoller
	serviceState map[string]*serviceSet
	closed       bool
	wg           sync.WaitGroup

	monitor modernMonitor
}

// Config returns a live property view using the client's default AppID.
func (c *ApolloClient) Config(ctx context.Context, namespace string) (Config, error) {
	return c.ConfigForApp(ctx, c.options.appID, namespace)
}

// ConfigForApp returns a live property view for an explicit AppID.
func (c *ApolloClient) ConfigForApp(ctx context.Context, appID, namespace string) (Config, error) {
	state, err := c.getState(ctx, appID, namespace, ParseConfigFileFormat(namespace))
	if err != nil {
		return nil, err
	}
	return state, nil
}

// ConfigFile returns a live raw-content view using the default AppID.
func (c *ApolloClient) ConfigFile(ctx context.Context, namespace string, format ConfigFileFormat) (ConfigFile, error) {
	return c.ConfigFileForApp(ctx, c.options.appID, namespace, format)
}

// ConfigFileForApp returns a live raw-content view for an explicit AppID.
func (c *ApolloClient) ConfigFileForApp(ctx context.Context, appID, namespace string, format ConfigFileFormat) (ConfigFile, error) {
	if format == "" {
		format = ParseConfigFileFormat(namespace)
	}
	namespace = namespaceForFormat(namespace, format)
	state, err := c.getState(ctx, appID, namespace, format)
	if err != nil {
		return nil, err
	}
	return (*modernConfigFile)(state), nil
}

// Load eagerly loads property namespaces and is the explicit v6 replacement
// for legacy MustStart. It stops at the first failed namespace.
func (c *ApolloClient) Load(ctx context.Context, namespaces ...string) error {
	for _, namespace := range namespaces {
		if _, err := c.Config(ctx, namespace); err != nil {
			return err
		}
	}
	return nil
}

func (c *ApolloClient) getState(ctx context.Context, appID, namespace string, format ConfigFileFormat) (*modernConfig, error) {
	if ctx == nil {
		return nil, errors.New("agollo: context is nil")
	}
	appID = strings.TrimSpace(appID)
	namespace = strings.TrimSpace(namespace)
	if appID == "" {
		return nil, errors.New("agollo: AppId is empty")
	}
	if namespace == "" {
		return nil, errors.New("agollo: namespace is empty")
	}
	if format == "" {
		format = ParseConfigFileFormat(namespace)
	}
	key := ConfigKey{AppID: appID, Cluster: c.options.cluster, Namespace: namespace, Format: format}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("agollo: client is closed")
	}
	state := c.states[key]
	if state == nil {
		state = newModernConfig(c, key)
		c.states[key] = state
	}
	c.mu.Unlock()

	if err := state.load(ctx); err != nil {
		return nil, err
	}
	if c.options.longPollEnabled && !c.options.localMode {
		c.ensurePoller(appID)
	}
	return state, nil
}

func (c *ApolloClient) ensurePoller(appID string) {
	c.mu.Lock()
	if c.closed || c.appPollers[appID] != nil {
		c.mu.Unlock()
		return
	}
	poller := &appPoller{client: c, appID: appID, wake: make(chan struct{}, 1)}
	c.appPollers[appID] = poller
	c.wg.Add(1)
	c.mu.Unlock()

	go func() {
		defer c.wg.Done()
		poller.run()
	}()
}

func (c *ApolloClient) Monitor() ClientMonitor { return &c.monitor }

func (c *ApolloClient) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.cancel()
	states := make([]*modernConfig, 0, len(c.states))
	for _, state := range c.states {
		states = append(states, state)
	}
	c.mu.Unlock()

	for _, state := range states {
		state.close()
	}
	c.wg.Wait()
	return nil
}

func (c *ApolloClient) goBackground(run func(context.Context)) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.wg.Add(1)
	c.mu.Unlock()
	go func() {
		defer c.wg.Done()
		run(c.ctx)
	}()
}

func (c *ApolloClient) beginSubscription() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	c.wg.Add(1)
	return true
}

// loadContext is canceled when either the operation context or the client
// lifecycle context ends. The watcher exits as soon as the returned cancel
// function is called, so individual loads do not leave goroutines behind.
func (c *ApolloClient) loadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	loadContext, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-c.ctx.Done():
			cancel()
		case <-loadContext.Done():
		}
	}()
	return loadContext, cancel
}

func normalizeURLs(urls []string) []string {
	result := make([]string, 0, len(urls))
	for _, rawURL := range urls {
		value := strings.TrimRight(strings.TrimSpace(rawURL), "/")
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func namespaceForFormat(namespace string, format ConfigFileFormat) string {
	namespace = strings.TrimSpace(namespace)
	suffix := "." + string(format)
	if strings.HasSuffix(strings.ToLower(namespace), suffix) {
		return namespace
	}
	return namespace + suffix
}

type modernMonitor struct {
	startedAt time.Time
	requests  atomic.Uint64
	failures  atomic.Uint64
	polls     atomic.Uint64
	pollFails atomic.Uint64
	drops     atomic.Uint64
	configs   atomic.Int64
}

func (m *modernMonitor) Snapshot() MonitorSnapshot {
	return MonitorSnapshot{
		StartedAt:       m.startedAt,
		ConfigCount:     int(m.configs.Load()),
		RemoteRequests:  m.requests.Load(),
		RemoteFailures:  m.failures.Load(),
		LongPolls:       m.polls.Load(),
		LongPollFailure: m.pollFails.Load(),
		ListenerDrops:   m.drops.Load(),
	}
}
