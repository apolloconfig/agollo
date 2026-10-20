package remote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
)

func TestNotificationMessagesFetch(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotModified, http.StatusBadRequest} {
		name := http.StatusText(status)
		if status == http.StatusBadRequest {
			name = "Malformed config response"
		}
		t.Run(name, func(t *testing.T) {
			expected := map[string]int64{"app+default+application": 102, "public+fallback+application": 100}
			requests := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/notifications/") {
					_ = json.NewEncoder(w).Encode([]config.Notification{{NamespaceName: "application", NotificationID: 102, Messages: &config.NotificationMessages{Details: expected}}})
					return
				}
				requests++
				var messages config.NotificationMessages
				if err := json.Unmarshal([]byte(r.URL.Query().Get("messages")), &messages); err != nil || !reflect.DeepEqual(messages.Details, expected) {
					t.Errorf("missing freshness information: %s (%v)", r.URL.RawQuery, err)
					w.WriteHeader(http.StatusNotModified) // A stale server would otherwise return 304.
					return
				}
				if r.URL.Query().Get("releaseKey") != "old-release" {
					t.Error("releaseKey not preserved")
				}
				if status == http.StatusBadRequest {
					_, _ = w.Write([]byte(`invalid JSON`))
					return
				}
				w.WriteHeader(status)
				if status == http.StatusOK {
					_, _ = w.Write([]byte(`{"appId":"app","cluster":"default","namespaceName":"application","releaseKey":"new-release","configurations":{"key":"new"}}`))
				}
			}))
			defer srv.Close()
			c := &config.AppConfig{AppID: "app", Cluster: "default", NamespaceName: "application", IP: srv.URL}
			c.Init()
			c.SetCurrentApolloConfig(&config.ApolloConnConfig{NamespaceName: "application", ReleaseKey: "old-release"})
			client := CreateAsyncApolloConfig()
			getConfig := func() config.AppConfig { return *c }
			got := client.Sync(getConfig)
			if status == http.StatusOK && (len(got) != 1 || got[0].ReleaseKey != "new-release") {
				t.Fatalf("expected new release, got %+v", got)
			}
			wantID := int64(102)
			if status != http.StatusOK {
				wantID = -1
			}
			if id := c.GetNotificationsMap().GetNotify("application"); id != wantID {
				t.Fatalf("notification ID = %d, want %d", id, wantID)
			}
			// A later fetch/retry still carries the observed versions, even after failure.
			server.SetNextTryConnTime(c.GetHost(), -1)
			_ = client.SyncWithNamespace("application", getConfig)
			if requests != 2 {
				t.Fatalf("got %d config requests, want 2", requests)
			}
		})
	}
}

func TestNotificationMessagesLegacyResponse(t *testing.T) {
	notifications, err := toApolloConfig([]byte(`[{"namespaceName":"application","notificationId":102}]`))
	if err != nil || len(notifications) != 1 || notifications[0].Messages != nil {
		t.Fatalf("legacy notification: %+v, %v", notifications, err)
	}
	c := &config.AppConfig{}
	c.Init()
	parsed, err := url.Parse(CreateAsyncApolloConfig().GetSyncURI(*c, "application"))
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := parsed.Query()["messages"]; exists {
		t.Fatal("empty messages must be omitted")
	}
}
