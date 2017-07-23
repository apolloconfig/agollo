package agollo

import (
	"testing"
	"time"
	"github.com/zouyx/agollo/test"
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

	test.Nil(t,err)
	test.NotNil(t,configRepository)

	t.Log(configRepository)
}
