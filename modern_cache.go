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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const modernCacheVersion = 1

type diskSnapshot struct {
	Version        int                    `json:"version"`
	AppID          string                 `json:"appId"`
	Cluster        string                 `json:"cluster"`
	Namespace      string                 `json:"namespaceName"`
	Format         ConfigFileFormat       `json:"format"`
	Values         map[string]interface{} `json:"configurations"`
	Content        string                 `json:"content"`
	ReleaseKey     string                 `json:"releaseKey"`
	NotificationID int64                  `json:"notificationId"`
}

func (c *ApolloClient) cacheFile(key ConfigKey) string {
	// Base64 keeps the filename one segment even if an AppId or namespace has
	// punctuation that would otherwise be interpreted as a path separator.
	identity := base64.RawURLEncoding.EncodeToString([]byte(key.String()))
	return filepath.Join(c.options.localCacheDir, identity+".agollo.json")
}

func (c *ApolloClient) legacyCacheFile(key ConfigKey) string {
	return filepath.Join(c.options.localCacheDir, key.AppID+"-"+key.Namespace+".json")
}

func (c *ApolloClient) persistLocalSnapshot(snapshot ConfigSnapshot) error {
	if c.options.localCacheDir == "" {
		return nil
	}
	disk := diskSnapshot{
		Version:        modernCacheVersion,
		AppID:          snapshot.Key.AppID,
		Cluster:        snapshot.Key.Cluster,
		Namespace:      snapshot.Key.Namespace,
		Format:         snapshot.Key.Format,
		Values:         snapshot.Values,
		Content:        snapshot.Content,
		ReleaseKey:     snapshot.ReleaseKey,
		NotificationID: snapshot.NotificationID,
	}
	body, err := json.Marshal(disk)
	if err != nil {
		return fmt.Errorf("agollo: encode local cache: %w", err)
	}
	file := c.cacheFile(snapshot.Key)
	temporary, err := os.CreateTemp(c.options.localCacheDir, ".agollo-*.tmp")
	if err != nil {
		return fmt.Errorf("agollo: create temporary cache file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := temporary.Write(body); err != nil {
		temporary.Close()
		return fmt.Errorf("agollo: write local cache: %w", err)
	}
	if err := temporary.Chmod(0o640); err != nil {
		temporary.Close()
		return fmt.Errorf("agollo: set local cache permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("agollo: close temporary cache file: %w", err)
	}
	if err := os.Rename(temporaryName, file); err != nil {
		return fmt.Errorf("agollo: atomically publish local cache: %w", err)
	}
	return nil
}

func (c *ApolloClient) loadLocalSnapshot(key ConfigKey) (ConfigSnapshot, error) {
	for _, file := range []string{c.cacheFile(key), c.legacyCacheFile(key)} {
		body, err := os.ReadFile(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return ConfigSnapshot{}, fmt.Errorf("agollo: read local cache: %w", err)
		}
		if snapshot, err := decodeDiskSnapshot(key, body); err == nil {
			return snapshot, nil
		}
	}
	return ConfigSnapshot{}, fmt.Errorf("agollo: no readable local cache for %s", key)
}

func decodeDiskSnapshot(key ConfigKey, body []byte) (ConfigSnapshot, error) {
	var disk diskSnapshot
	if err := json.Unmarshal(body, &disk); err != nil {
		return ConfigSnapshot{}, err
	}
	// agollo legacy cache has the same JSON field names but no version/format.
	if disk.AppID != "" && disk.AppID != key.AppID {
		return ConfigSnapshot{}, errors.New("local cache AppId does not match requested AppId")
	}
	if disk.Namespace != "" && disk.Namespace != key.Namespace {
		return ConfigSnapshot{}, errors.New("local cache namespace does not match requested namespace")
	}
	if disk.Cluster != "" && disk.Cluster != key.Cluster {
		return ConfigSnapshot{}, errors.New("local cache cluster does not match requested cluster")
	}
	values := cloneValues(disk.Values)
	if values == nil {
		values = map[string]interface{}{}
	}
	content := disk.Content
	if content == "" {
		content = extractContent(values, key.Format)
	}
	if key.Format.IsPropertiesCompatible() && content != "" && len(values) == 1 {
		if parsed, err := parsePropertiesCompatibleContent(key.Format, content); err == nil && len(parsed) > 0 {
			values = parsed
		}
	}
	return ConfigSnapshot{
		Key:            key,
		Values:         values,
		Content:        content,
		ReleaseKey:     strings.TrimSpace(disk.ReleaseKey),
		NotificationID: disk.NotificationID,
		Source:         ConfigSourceLocalFile,
		UpdatedAt:      now(),
	}, nil
}

var now = func() time.Time { return time.Now() }
