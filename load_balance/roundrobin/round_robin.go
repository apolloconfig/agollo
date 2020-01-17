package roundrobin

import (
	"github.com/zouyx/agollo/v2/env/config"
	"github.com/zouyx/agollo/v2/load_balance"
	"sync"
)

func init() {
	load_balance.SetLoadBalance(&RoundRobin{})
}

//RoundRobin 轮询调度
type RoundRobin struct {
}

//Load 负载均衡
func (r *RoundRobin) Load(servers *sync.Map) *config.ServerInfo {
	var returnServer *config.ServerInfo
	servers.Range(func(k, v interface{}) bool {
		server := v.(*config.ServerInfo)
		// if some node has down then select next node
		if server.IsDown {
			return true
		}
		returnServer = server
		return false
	})
	return returnServer
}
