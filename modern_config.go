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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type modernConfig struct {
	client *ApolloClient
	key    ConfigKey

	snapshot atomic.Value // *ConfigSnapshot

	loadMu  sync.Mutex
	loaded  bool
	loadErr error
	closed  atomic.Bool

	listenerMu    sync.Mutex
	listeners     map[uint64]*configSubscription
	fileListeners map[uint64]*fileSubscription
	nextListener  uint64
}

func newModernConfig(client *ApolloClient, key ConfigKey) *modernConfig {
	state := &modernConfig{
		client:        client,
		key:           key,
		listeners:     make(map[uint64]*configSubscription),
		fileListeners: make(map[uint64]*fileSubscription),
	}
	state.snapshot.Store(&ConfigSnapshot{
		Key:       key,
		Values:    map[string]interface{}{},
		Source:    ConfigSourceNone,
		UpdatedAt: time.Now(),
	})
	return state
}

func (c *modernConfig) Key() ConfigKey { return c.key }

func (c *modernConfig) Snapshot() ConfigSnapshot {
	return c.current().Clone()
}

func (c *modernConfig) current() *ConfigSnapshot {
	return c.snapshot.Load().(*ConfigSnapshot)
}

func (c *modernConfig) Lookup(key string) (string, bool) {
	value, ok := c.current().Values[key]
	if !ok || value == nil {
		return "", false
	}
	return stringify(value)
}

func (c *modernConfig) String(key, defaultValue string) string {
	if value, ok := c.Lookup(key); ok {
		return value
	}
	return defaultValue
}

func (c *modernConfig) Int(key string, defaultValue int) int {
	value, ok := c.Lookup(key)
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (c *modernConfig) Int64(key string, defaultValue int64) int64 {
	value, ok := c.Lookup(key)
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (c *modernConfig) Float64(key string, defaultValue float64) float64 {
	value, ok := c.Lookup(key)
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (c *modernConfig) Bool(key string, defaultValue bool) bool {
	value, ok := c.Lookup(key)
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (c *modernConfig) Duration(key string, defaultValue time.Duration) time.Duration {
	value, ok := c.Lookup(key)
	if !ok {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (c *modernConfig) StringSlice(key string, defaultValue []string) []string {
	value, exists := c.current().Values[key]
	if !exists {
		return append([]string(nil), defaultValue...)
	}
	result, ok := stringSlice(value)
	if !ok {
		return append([]string(nil), defaultValue...)
	}
	return result
}

func (c *modernConfig) IntSlice(key string, defaultValue []int) []int {
	value, exists := c.current().Values[key]
	if !exists {
		return append([]int(nil), defaultValue...)
	}
	result, ok := intSlice(value)
	if !ok {
		return append([]int(nil), defaultValue...)
	}
	return result
}

func (c *modernConfig) Keys() []string { return sortedKeys(c.current().Values) }

func (c *modernConfig) Source() ConfigSourceType { return c.current().Source }

func (c *modernConfig) Subscribe(listener ConfigChangeHandler, options ...SubscribeOption) func() {
	if listener == nil {
		return func() {}
	}
	settings := subscribeOptions{}
	for _, option := range options {
		if option != nil {
			option(&settings)
		}
	}

	c.listenerMu.Lock()
	if c.closed.Load() {
		c.listenerMu.Unlock()
		return func() {}
	}
	if !c.client.beginSubscription() {
		c.listenerMu.Unlock()
		return func() {}
	}
	id := c.nextListener
	c.nextListener++
	subscription := newConfigSubscription(listener, settings, c.client.options.listenerQueueSize, &c.client.monitor, &c.client.wg)
	c.listeners[id] = subscription
	c.listenerMu.Unlock()

	return func() {
		c.listenerMu.Lock()
		current := c.listeners[id]
		delete(c.listeners, id)
		c.listenerMu.Unlock()
		if current != nil {
			current.close()
		}
	}
}

func (c *modernConfig) load(ctx context.Context) error {
	c.loadMu.Lock()
	defer c.loadMu.Unlock()
	if c.closed.Load() {
		return fmt.Errorf("agollo: config %s is closed", c.key)
	}
	if c.loaded {
		return c.loadErr
	}
	if err := c.client.loadConfig(ctx, c); err != nil {
		c.loadErr = err
		return err
	}
	c.loaded = true
	c.loadErr = nil
	c.client.monitor.configs.Add(1)
	return nil
}

func (c *modernConfig) reload(ctx context.Context, notification notification) error {
	c.loadMu.Lock()
	defer c.loadMu.Unlock()
	if c.closed.Load() {
		return nil
	}
	if err := c.client.loadConfigWithNotification(ctx, c, notification); err != nil {
		return err
	}
	c.loaded = true
	c.loadErr = nil
	return nil
}

// acknowledgeNotification advances only the long-poll cursor. A 304 response
// has already confirmed the notification, but must not produce a config event.
func (c *modernConfig) acknowledgeNotification(id int64) {
	if id <= 0 {
		return
	}
	current := c.current()
	if id <= current.NotificationID {
		return
	}
	next := *current
	next.NotificationID = id
	c.snapshot.Store(&next)
}

func (c *modernConfig) publish(next ConfigSnapshot) {
	previous := c.current()
	next.Key = c.key
	next.Values = cloneValues(next.Values)
	if next.Values == nil {
		next.Values = map[string]interface{}{}
	}
	if next.UpdatedAt.IsZero() {
		next.UpdatedAt = time.Now()
	}
	c.snapshot.Store(&next)

	changes := diffValues(previous.Values, next.Values)
	contentChanged := previous.Content != next.Content
	if len(changes) == 0 && !contentChanged {
		return
	}
	if len(changes) > 0 {
		event := ConfigChangeEvent{
			Key:        c.key,
			Changes:    changes,
			ReleaseKey: next.ReleaseKey,
			Source:     next.Source,
			OccurredAt: next.UpdatedAt,
		}
		c.publishConfigChange(event)
	}
	if contentChanged {
		event := ConfigFileChangeEvent{
			Key:        c.key,
			OldContent: previous.Content,
			NewContent: next.Content,
			ReleaseKey: next.ReleaseKey,
			Source:     next.Source,
			OccurredAt: next.UpdatedAt,
		}
		c.publishFileChange(event)
	}
}

func (c *modernConfig) publishConfigChange(event ConfigChangeEvent) {
	c.listenerMu.Lock()
	subscriptions := make([]*configSubscription, 0, len(c.listeners))
	for _, subscription := range c.listeners {
		subscriptions = append(subscriptions, subscription)
	}
	c.listenerMu.Unlock()
	for _, subscription := range subscriptions {
		if subscription.options.matches(event.Changes) {
			subscription.offer(event)
		}
	}
}

func (c *modernConfig) publishFileChange(event ConfigFileChangeEvent) {
	c.listenerMu.Lock()
	subscriptions := make([]*fileSubscription, 0, len(c.fileListeners))
	for _, subscription := range c.fileListeners {
		subscriptions = append(subscriptions, subscription)
	}
	c.listenerMu.Unlock()
	for _, subscription := range subscriptions {
		subscription.offer(event)
	}
}

func (c *modernConfig) close() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	c.listenerMu.Lock()
	listeners := c.listeners
	fileListeners := c.fileListeners
	c.listeners = make(map[uint64]*configSubscription)
	c.fileListeners = make(map[uint64]*fileSubscription)
	c.listenerMu.Unlock()
	for _, listener := range listeners {
		listener.close()
	}
	for _, listener := range fileListeners {
		listener.close()
	}
}

type modernConfigFile modernConfig

func (c *modernConfigFile) config() *modernConfig { return (*modernConfig)(c) }

func (c *modernConfigFile) Key() ConfigKey { return c.config().Key() }

func (c *modernConfigFile) Content() string { return c.config().current().Content }

func (c *modernConfigFile) HasContent() bool { return c.Content() != "" }

func (c *modernConfigFile) Format() ConfigFileFormat { return c.config().key.Format }

func (c *modernConfigFile) Source() ConfigSourceType { return c.config().Source() }

func (c *modernConfigFile) AsMap() (map[string]interface{}, bool) {
	if !c.Format().IsPropertiesCompatible() {
		return nil, false
	}
	return cloneValues(c.config().current().Values), true
}

func (c *modernConfigFile) Subscribe(listener ConfigFileChangeHandler) func() {
	config := c.config()
	if listener == nil {
		return func() {}
	}
	config.listenerMu.Lock()
	if config.closed.Load() {
		config.listenerMu.Unlock()
		return func() {}
	}
	if !config.client.beginSubscription() {
		config.listenerMu.Unlock()
		return func() {}
	}
	id := config.nextListener
	config.nextListener++
	subscription := newFileSubscription(listener, config.client.options.listenerQueueSize, &config.client.monitor, &config.client.wg)
	config.fileListeners[id] = subscription
	config.listenerMu.Unlock()
	return func() {
		config.listenerMu.Lock()
		current := config.fileListeners[id]
		delete(config.fileListeners, id)
		config.listenerMu.Unlock()
		if current != nil {
			current.close()
		}
	}
}

func stringify(value interface{}) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case fmt.Stringer:
		return typed.String(), true
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(typed), true
	default:
		return "", false
	}
}

func diffValues(previous, next map[string]interface{}) map[string]PropertyChange {
	changes := make(map[string]PropertyChange)
	for key, previousValue := range previous {
		nextValue, exists := next[key]
		if !exists {
			changes[key] = PropertyChange{OldValue: cloneValue(previousValue), Type: ChangeDeleted}
			continue
		}
		if !reflect.DeepEqual(previousValue, nextValue) {
			changes[key] = PropertyChange{OldValue: cloneValue(previousValue), NewValue: cloneValue(nextValue), Type: ChangeModified}
		}
	}
	for key, nextValue := range next {
		if _, exists := previous[key]; !exists {
			changes[key] = PropertyChange{NewValue: cloneValue(nextValue), Type: ChangeAdded}
		}
	}
	return changes
}

type configSubscription struct {
	listener ConfigChangeHandler
	options  subscribeOptions
	queue    chan ConfigChangeEvent
	done     chan struct{}
	once     sync.Once
	monitor  *modernMonitor
	wg       *sync.WaitGroup
}

func newConfigSubscription(listener ConfigChangeHandler, options subscribeOptions, size int, monitor *modernMonitor, wg *sync.WaitGroup) *configSubscription {
	subscription := &configSubscription{listener: listener, options: options, queue: make(chan ConfigChangeEvent, size), done: make(chan struct{}), monitor: monitor, wg: wg}
	go subscription.run()
	return subscription
}

func (s *configSubscription) offer(event ConfigChangeEvent) {
	select {
	case <-s.done:
		return
	default:
	}
	select {
	case s.queue <- event:
		return
	default:
	}

	// Keep the latest namespace event rather than creating unbounded goroutines.
	select {
	case <-s.queue:
	default:
	}
	select {
	case s.queue <- event:
		s.monitor.drops.Add(1)
	case <-s.done:
	}
}

func (s *configSubscription) run() {
	defer s.wg.Done()
	for {
		select {
		case <-s.done:
			return
		case event := <-s.queue:
			func() {
				defer func() { _ = recover() }()
				s.listener(event)
			}()
		}
	}
}

func (s *configSubscription) close() { s.once.Do(func() { close(s.done) }) }

type fileSubscription struct {
	listener ConfigFileChangeHandler
	queue    chan ConfigFileChangeEvent
	done     chan struct{}
	once     sync.Once
	monitor  *modernMonitor
	wg       *sync.WaitGroup
}

func newFileSubscription(listener ConfigFileChangeHandler, size int, monitor *modernMonitor, wg *sync.WaitGroup) *fileSubscription {
	subscription := &fileSubscription{listener: listener, queue: make(chan ConfigFileChangeEvent, size), done: make(chan struct{}), monitor: monitor, wg: wg}
	go subscription.run()
	return subscription
}

func (s *fileSubscription) offer(event ConfigFileChangeEvent) {
	select {
	case <-s.done:
		return
	default:
	}
	select {
	case s.queue <- event:
		return
	default:
	}
	select {
	case <-s.queue:
	default:
	}
	select {
	case s.queue <- event:
		s.monitor.drops.Add(1)
	case <-s.done:
	}
}

func (s *fileSubscription) run() {
	defer s.wg.Done()
	for {
		select {
		case <-s.done:
			return
		case event := <-s.queue:
			func() {
				defer func() { _ = recover() }()
				s.listener(event)
			}()
		}
	}
}

func (s *fileSubscription) close() { s.once.Do(func() { close(s.done) }) }

var _ Config = (*modernConfig)(nil)
var _ ConfigFile = (*modernConfigFile)(nil)

func waitWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	return normalizeURLs(parts)
}
