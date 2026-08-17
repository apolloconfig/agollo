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
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// notification is Apollo's notification-v2 payload plus the optional remote
// message map used by incremental configuration sync.
type notification struct {
	NamespaceName string           `json:"namespaceName"`
	ID            int64            `json:"notificationId"`
	Messages      map[string]int64 `json:"messages"`
}

type appPoller struct {
	client *ApolloClient
	appID  string
	wake   chan struct{}
}

func (p *appPoller) run() {
	if delay := p.client.options.pollInitialDelay; delay > 0 {
		if waitWithContext(p.client.ctx, delay) != nil {
			return
		}
	}
	attempt := 0
	for {
		if err := p.poll(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			p.client.monitor.pollFails.Add(1)
			if waitWithContext(p.client.ctx, retryDelay(p.client.options.retryMin, p.client.options.retryMax, attempt)) != nil {
				return
			}
			attempt++
			continue
		}
		attempt = 0
	}
}

func (p *appPoller) poll() error {
	states := p.client.statesForAppID(p.appID)
	if len(states) == 0 {
		select {
		case <-p.client.ctx.Done():
			return p.client.ctx.Err()
		case <-p.wake:
			return nil
		case <-time.After(time.Second):
			return nil
		}
	}

	services, err := p.client.configServices(p.client.ctx, p.appID)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return errors.New("agollo: no Config Service for long poll")
	}
	serviceURL, err := p.client.selectConfigService(p.appID, services)
	if err != nil {
		return err
	}
	if serviceURL == "" {
		return errors.New("agollo: no Config Service for long poll")
	}
	endpoint, err := p.notificationsURL(serviceURL, states)
	if err != nil {
		return err
	}
	body, err := p.client.getJSON(p.client.ctx, endpoint, p.appID, 90*time.Second, http.Header{})
	p.client.monitor.polls.Add(1)
	if errors.Is(err, errNotModified) {
		return nil
	}
	if err != nil {
		return err
	}
	var notifications []notification
	if err := json.Unmarshal(body, &notifications); err != nil {
		return fmt.Errorf("agollo: decode notifications response: %w", err)
	}
	for _, notice := range notifications {
		if notice.NamespaceName == "" {
			continue
		}
		for _, state := range states {
			if !notificationMatchesNamespace(notice.NamespaceName, state.key.Namespace) {
				continue
			}
			if err := state.reload(p.client.ctx, notice); err != nil && !errors.Is(err, errNotModified) {
				return err
			}
		}
	}
	return nil
}

func (p *appPoller) notificationsURL(serviceURL string, states []*modernConfig) (string, error) {
	base, err := url.Parse(serviceURL)
	if err != nil {
		return "", fmt.Errorf("agollo: invalid Config Service URL %q: %w", serviceURL, err)
	}
	base.Path = path.Join(base.Path, "notifications", "v2")
	notifications := make([]notification, 0, len(states))
	seen := make(map[string]struct{}, len(states))
	for _, state := range states {
		namespace := state.key.Namespace
		if _, exists := seen[namespace]; exists {
			continue
		}
		seen[namespace] = struct{}{}
		notifications = append(notifications, notification{NamespaceName: namespace, ID: state.current().NotificationID})
	}
	encoded, err := json.Marshal(notifications)
	if err != nil {
		return "", fmt.Errorf("agollo: encode notifications request: %w", err)
	}
	query := base.Query()
	query.Set("appId", p.appID)
	query.Set("cluster", p.client.options.cluster)
	query.Set("notifications", string(encoded))
	if p.client.options.dataCenter != "" {
		query.Set("dataCenter", p.client.options.dataCenter)
	}
	if p.client.options.localIP != "" {
		query.Set("ip", p.client.options.localIP)
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (c *ApolloClient) statesForAppID(appID string) []*modernConfig {
	c.mu.Lock()
	defer c.mu.Unlock()
	states := make([]*modernConfig, 0)
	for key, state := range c.states {
		if key.AppID == appID && !state.closed.Load() {
			states = append(states, state)
		}
	}
	return states
}

func notificationMatchesNamespace(notified, requested string) bool {
	if notified == requested {
		return true
	}
	return strings.TrimSuffix(requested, ".properties") == notified
}
