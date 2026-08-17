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
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultModernCluster = "default"

// ClientOption configures a ConfigClient. Options are applied before any
// network request or goroutine is created.
type ClientOption func(*clientOptions) error

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

// WithAppID sets the default AppId. It is required unless every lookup uses
// GetConfigFor/GetConfigFileFor with an explicit AppId.
func WithAppID(appID string) ClientOption {
	return func(options *clientOptions) error {
		options.appID = strings.TrimSpace(appID)
		return nil
	}
}

// WithCluster sets the Apollo cluster. Empty values retain Apollo's default.
func WithCluster(cluster string) ClientOption {
	return func(options *clientOptions) error {
		if value := strings.TrimSpace(cluster); value != "" {
			options.cluster = value
		}
		return nil
	}
}

// WithConfigServiceURLs bypasses Meta Server discovery and uses the supplied
// Config Service addresses in order. Each URL must include a scheme.
func WithConfigServiceURLs(urls ...string) ClientOption {
	return func(options *clientOptions) error {
		options.configServiceURLs = normalizeURLs(urls)
		if len(options.configServiceURLs) == 0 {
			return errors.New("agollo: at least one non-empty config service URL is required")
		}
		return nil
	}
}

// WithMetaServerURL configures the Meta Server used for Config Service discovery.
func WithMetaServerURL(url string) ClientOption {
	return func(options *clientOptions) error {
		options.metaServerURL = strings.TrimRight(strings.TrimSpace(url), "/")
		return nil
	}
}

// WithAccessKeySecret configures the default AppId access key secret.
func WithAccessKeySecret(secret string) ClientOption {
	return func(options *clientOptions) error {
		options.secret = secret
		return nil
	}
}

// WithAppIDAccessKeySecret configures an AppId-specific access key secret.
func WithAppIDAccessKeySecret(appID, secret string) ClientOption {
	return func(options *clientOptions) error {
		appID = strings.TrimSpace(appID)
		if appID == "" {
			return errors.New("agollo: appId for access key secret is empty")
		}
		if options.secretsByAppID == nil {
			options.secretsByAppID = make(map[string]string)
		}
		options.secretsByAppID[appID] = secret
		return nil
	}
}

// WithLabel configures Apollo gray-release label selection.
func WithLabel(label string) ClientOption {
	return func(options *clientOptions) error {
		options.label = strings.TrimSpace(label)
		return nil
	}
}

// WithDataCenter configures Apollo dataCenter routing for fetch and long poll.
func WithDataCenter(dataCenter string) ClientOption {
	return func(options *clientOptions) error {
		options.dataCenter = strings.TrimSpace(dataCenter)
		return nil
	}
}

// WithLocalIP configures the client IP reported to Apollo. If omitted, no ip
// query parameter is sent by the modern client.
func WithLocalIP(ip string) ClientOption {
	return func(options *clientOptions) error {
		options.localIP = strings.TrimSpace(ip)
		return nil
	}
}

// WithHTTPClient supplies the HTTP client used by discovery, config fetches,
// and long polling. Its Transport must be safe for concurrent use.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(options *clientOptions) error {
		if client == nil {
			return errors.New("agollo: HTTP client is nil")
		}
		options.httpClient = client
		return nil
	}
}

// WithLocalCacheDir enables durable local cache fallback. The implementation
// reads both the modern cache file and agollo's legacy JSON cache format.
func WithLocalCacheDir(dir string) ClientOption {
	return func(options *clientOptions) error {
		options.localCacheDir = strings.TrimSpace(dir)
		return nil
	}
}

// WithConfigMapStore enables the optional Kubernetes ConfigMap fallback.
// ConfigMap is consulted only after remote and local-file sources fail.
func WithConfigMapStore(store ConfigMapStore) ClientOption {
	return func(options *clientOptions) error {
		if store == nil {
			return errors.New("agollo: ConfigMap store is nil")
		}
		options.configMapStore = store
		return nil
	}
}

// WithLocalMode disables all remote traffic. Configurations can only be loaded
// from the local cache directory.
func WithLocalMode() ClientOption {
	return func(options *clientOptions) error {
		options.localMode = true
		options.longPollEnabled = false
		return nil
	}
}

// WithoutLongPoll is useful for command line jobs and deterministic tests.
func WithoutLongPoll() ClientOption {
	return func(options *clientOptions) error {
		options.longPollEnabled = false
		return nil
	}
}

// WithLongPollInitialDelay delays long polling after the first successful
// namespace load.
func WithLongPollInitialDelay(delay time.Duration) ClientOption {
	return func(options *clientOptions) error {
		if delay < 0 {
			return errors.New("agollo: long poll initial delay cannot be negative")
		}
		options.pollInitialDelay = delay
		return nil
	}
}

// WithRetryBackoff configures the bounded exponential retry range.
func WithRetryBackoff(min, max time.Duration) ClientOption {
	return func(options *clientOptions) error {
		if min <= 0 || max < min {
			return errors.New("agollo: invalid retry backoff range")
		}
		options.retryMin = min
		options.retryMax = max
		return nil
	}
}

// WithListenerQueueSize limits outstanding events per listener. On overflow,
// intermediate events are coalesced to the latest event and recorded by Monitor.
func WithListenerQueueSize(size int) ClientOption {
	return func(options *clientOptions) error {
		if size < 1 {
			return errors.New("agollo: listener queue size must be positive")
		}
		options.listenerQueueSize = size
		return nil
	}
}

// NewClient creates an instance-scoped client. It does not perform network I/O
// until a namespace is requested.
func NewClient(ctx context.Context, options ...ClientOption) (ConfigClient, error) {
	if ctx == nil {
		return nil, errors.New("agollo: context is nil")
	}

	config := defaultClientOptions()
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	if config.localMode && config.localCacheDir == "" {
		return nil, errors.New("agollo: local mode requires WithLocalCacheDir")
	}
	if !config.localMode && len(config.configServiceURLs) == 0 && config.metaServerURL == "" {
		return nil, errors.New("agollo: configure Config Service URLs or a Meta Server URL")
	}
	if config.localCacheDir != "" {
		if err := os.MkdirAll(config.localCacheDir, 0o750); err != nil {
			return nil, fmt.Errorf("agollo: create local cache directory: %w", err)
		}
	}

	clientContext, cancel := context.WithCancel(ctx)
	client := &modernClient{
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

type modernClient struct {
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

func (c *modernClient) GetConfig(ctx context.Context, namespace string) (Config, error) {
	return c.GetConfigFor(ctx, c.options.appID, namespace)
}

func (c *modernClient) GetConfigFor(ctx context.Context, appID, namespace string) (Config, error) {
	state, err := c.getState(ctx, appID, namespace, ParseConfigFileFormat(namespace))
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (c *modernClient) GetConfigFile(ctx context.Context, namespace string, format ConfigFileFormat) (ConfigFile, error) {
	return c.GetConfigFileFor(ctx, c.options.appID, namespace, format)
}

func (c *modernClient) GetConfigFileFor(ctx context.Context, appID, namespace string, format ConfigFileFormat) (ConfigFile, error) {
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

func (c *modernClient) getState(ctx context.Context, appID, namespace string, format ConfigFileFormat) (*modernConfig, error) {
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

func (c *modernClient) ensurePoller(appID string) {
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

func (c *modernClient) Monitor() ClientMonitor { return &c.monitor }

func (c *modernClient) Close() error {
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

func (c *modernClient) goBackground(run func(context.Context)) {
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
