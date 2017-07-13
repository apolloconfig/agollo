package timer

import (
	"testing"
	"github.com/zouyx/agollo/component"
	"time"
	"github.com/zouyx/agollo/config"
	"github.com/zouyx/agollo/repository"
	"github.com/zouyx/agollo/test"
)

func TestInitRefreshInterval(t *testing.T) {
	config.REFRESH_INTERVAL=1*time.Second

	var c component.AbsComponent
	c=&AutoRefreshConfigComponent{}
	c.Start()
}

func TestSyncConfigServices(t *testing.T) {
	err:=syncConfigServices()

	configRepository:=repository.GetCurrentApolloConfig()

	test.Nil(t,err)
	test.NotNil(t,configRepository)

	t.Log(configRepository)
}
