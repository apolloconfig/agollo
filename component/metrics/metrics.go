package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NamespaceVersionGauge 当前 namespace 版本、notifyId，值为 status，0 历史版本，1 当前版本
	NamespaceVersionGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "namespace_version",
		Subsystem: "sdk",
		Namespace: "apollo",
	}, []string{"appid", "req_cluster", "use_cluster", "namespace", "ip", "version"})

	// LatestCheckGauge namespace 最近检查监控，值为触发时间戳
	LatestCheckGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "latest_check",
		Subsystem: "sdk",
		Namespace: "apollo",
	}, []string{"appid", "req_cluster", "ip", "namespaces", "addr"})

	// OnchangeGauge 触发 onchange 监控, 值为版本
	OnchangeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "onchange",
		Subsystem: "sdk",
		Namespace: "apollo",
	}, []string{"appid", "use_cluster", "namespace", "ip"})
)

var m = make(map[string]string)

func GetVersionByNamespace(appid, cluster, namespace string) string {
	return m[appid+cluster+namespace]
}

func SetVersionByNamespace(appid, cluster, namespace, version string) {
	m[appid+cluster+namespace] = version
}
