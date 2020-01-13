package load_balance

import (
	"github.com/zouyx/agollo/v2/env/config"
	"sync"
)

var defaultLoadBalance LoadBalance

type LoadBalance interface {
	Load(servers *sync.Map) *config.ServerInfo
}

func SetLoadBalance(loadBalance LoadBalance) {
	defaultLoadBalance = loadBalance
}

func GetLoadBalance() LoadBalance {
	return defaultLoadBalance
}
