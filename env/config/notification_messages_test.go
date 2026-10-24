package config

import (
	"sync"
	"testing"
)

func TestNotificationMessagesMergeAndIsolation(t *testing.T) {
	c := &AppConfig{NamespaceName: "application"}
	c.Init()
	n := c.GetNotificationsMap()
	incoming := &NotificationMessages{Details: map[string]int64{"app+default+application": 102, "public+default+application": 99}}
	n.RecordMessageIDs("application", incoming)
	incoming.Details["app+default+application"] = 1
	n.RecordMessageIDs("application", incoming)
	snapshot := n.GetMessageIDs("application")
	if snapshot.Details["app+default+application"] != 102 || snapshot.Details["public+default+application"] != 99 {
		t.Fatalf("unexpected messages: %+v", snapshot)
	}
	snapshot.Details["app+default+application"] = 0
	if n.GetMessageIDs("application").Details["app+default+application"] != 102 {
		t.Fatal("snapshot mutated shared state")
	}
	if n.GetNotify("application") != -1 {
		t.Fatal("observing messages acknowledged notification")
	}
	if n.GetMessageIDs("other") != nil {
		t.Fatal("messages leaked across namespaces")
	}
	other := &AppConfig{}
	other.Init()
	if other.GetNotificationsMap().GetMessageIDs("application") != nil {
		t.Fatal("messages leaked across clients")
	}
	n.RecordMessageIDs("application", nil)
}

func TestNotificationMessagesConcurrentMerge(t *testing.T) {
	c := &AppConfig{}
	c.Init()
	var wg sync.WaitGroup
	for i := int64(1); i <= 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			c.GetNotificationsMap().RecordMessageIDs("application", &NotificationMessages{Details: map[string]int64{"app+default+application": id}})
			_ = c.GetNotificationsMap().GetMessageIDs("application")
		}(i)
	}
	wg.Wait()
	if got := c.GetNotificationsMap().GetMessageIDs("application").Details["app+default+application"]; got != 100 {
		t.Fatalf("got %d, want 100", got)
	}
}
