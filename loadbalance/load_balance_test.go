package loadbalance

import (
	"sync"
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/env/config"
)

type TestLoadBalance struct {
}

//Load 负载均衡
func (r *TestLoadBalance) Load(servers *sync.Map) *config.ServerInfo {
	return nil
}

func TestSetLoadBalance(t *testing.T) {
	SetLoadBalance(&TestLoadBalance{})

	balance := GetLoadBalance()

	b := balance.(*TestLoadBalance)
	Assert(t, b, NotNilVal())
}
