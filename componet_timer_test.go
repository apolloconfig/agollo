package agollo

import (
	"testing"
	"time"
)

func TestInitRefreshInterval(t *testing.T) {
	refresh_interval=1*time.Second

	var c AbsComponent
	c=&AutoRefreshConfigComponent{}
	c.Start()
}

func TestSyncConfigServices_1(t *testing.T) {
	err:=syncConfigServices()

	configRepository:=GetCurrentApolloConfig()

	Nil(t,err)
	NotNil(t,configRepository)

	t.Log(configRepository)
}
