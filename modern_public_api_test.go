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

package agollo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apolloconfig/agollo/v6"
)

// This test deliberately uses only exported identifiers. It is the compile
// contract for the API shown in README and the migration guide.
func TestPublicApolloClientAPI(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.URL.Path, "/")
		if len(parts) != 5 || parts[1] != "configs" {
			t.Errorf("unexpected request path: %s", request.URL.Path)
			http.Error(writer, "unexpected request path", http.StatusNotFound)
			return
		}
		configurations := map[string]interface{}{"owner": parts[2]}
		if strings.HasSuffix(parts[4], ".yaml") {
			configurations = map[string]interface{}{"content": "server:\n  port: 8080\n"}
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(map[string]interface{}{
			"releaseKey": "r1", "configurations": configurations,
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := agollo.NewClient(context.Background(), agollo.ClientOptions{
		AppID:              "orders",
		Cluster:            "default",
		ConfigServices:     []string{server.URL},
		DisableLongPolling: true,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	config, err := client.Config(context.Background(), "application")
	if err != nil || config.String("owner", "") != "orders" {
		t.Fatalf("Config() = %v, %v", config, err)
	}
	other, err := client.ConfigForApp(context.Background(), "payments", "application")
	if err != nil || other.String("owner", "") != "payments" {
		t.Fatalf("ConfigForApp() = %v, %v", other, err)
	}
	file, err := client.ConfigFile(context.Background(), "application", agollo.ConfigFileFormatYAML)
	if err != nil || !strings.Contains(file.Content(), "8080") {
		t.Fatalf("ConfigFile() = %v, %v", file, err)
	}
}
