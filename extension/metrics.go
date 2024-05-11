package extension

import (
	"sync"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/component/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

/**
 * Metrics 指标上报模块
 */

var (
	globalMetrics *prometheus.Registry
	once          sync.Once
)

// GetMetricsRegister 获取 metrics register
func GetMetricsRegister() *prometheus.Registry {
	return globalMetrics
}

// SetMetricsRegister 甚至 metrics register
func SetMetricsRegister(register *prometheus.Registry) {
	globalMetrics = register
}

func InitMetrics() {
	once.Do(func() {
		register := GetMetricsRegister()
		if register == nil {
			log.Warnf("not setting metrics register")
			return
		}

		// namespace 版本监控
		if err := register.Register(metrics.NamespaceVersionGauge); err != nil {
			log.Errorf("register namespace_version metrics fail, error: %v", err)
		}

		// 最近更新内容监控
		if err := register.Register(metrics.LatestCheckGauge); err != nil {
			log.Errorf("register namespace_version metrics fail, error: %v", err)
		}

		// onchange 监控
		if err := register.Register(metrics.OnchangeGauge); err != nil {
			log.Errorf("register namespace_version metrics fail, error: %v", err)
		}
	})
}
