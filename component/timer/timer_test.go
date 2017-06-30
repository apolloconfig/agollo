package timer

import (
	"testing"
	"github.com/zouyx/agollo/component"
	"time"
	"github.com/zouyx/agollo/config"
	"github.com/zouyx/agollo/repository"
)

func TestInitRefreshInterval(t *testing.T) {
	config.REFRESH_INTERVAL=1*time.Second

	var c component.AbsComponent
	c=&AutoRefreshConfigComponent{}
	c.Start()
}

func TestUpdateConfigServices(t *testing.T) {
	updateConfigServices()

	configRepository:=repository.GetConfig()

	t.Log(configRepository)
}
