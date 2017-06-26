package timer

import (
	"testing"
	"github.com/zouyx/agollo/component"
	"time"
	"github.com/zouyx/agollo/config"
)

func TestInitRefreshInterval(t *testing.T) {
	config.REFRESH_INTERVAL=1
	config.REFRESH_INTERVAL_TIME_UNIT=time.Second

	var c component.AbsComponent
	c=&AutoRefreshConfigComponent{}
	c.Start()
}
