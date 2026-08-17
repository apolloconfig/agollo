# agollo Java Client 能力迁移：首轮实现与测试说明

> 分支：`feature/java-client-parity-refactor`
> 基线：agollo `2ced0e9`；Apollo Java Client `023217c`
> 状态：v6 可运行内核；公开文档和新能力以 `ApolloClient` 为准，发布前仍须完成生态适配门禁。

## 本次交付

v6 使用 `github.com/apolloconfig/agollo/v6` 模块路径，并以实例级 `ApolloClient` 作为公开 API。构造函数返回具体类型，启动参数集中在可发现、可校验的 `ClientOptions` 中：

```go
client, err := agollo.NewClient(ctx, agollo.ClientOptions{
    AppID:      "orders",
    Cluster:    "default",
    MetaServer: "http://apollo-meta:8080",
    CacheDir:   "/var/lib/orders/apollo",
})
if err != nil { /* handle error */ }
defer client.Close()

cfg, err := client.Config(ctx, "application")
port := cfg.Int("server.port", 8080)

file, err := client.ConfigFile(ctx, "application", agollo.ConfigFileFormatYAML)
cancel := cfg.Subscribe(onChange, agollo.WithInterestedKeyPrefixes("db."))
defer cancel()
_ = port
_ = file
```

已实现的 Java Client 对齐能力：

- `ConfigKey{AppID, Cluster, Namespace, Format}` 作为隔离边界；单个客户端可用 `ConfigForApp` / `ConfigFileForApp` 获取多个 AppId 的配置。
- `Config` 的并发安全实时快照、`Lookup`、字符串/int/int64/float/bool/duration/字符串切片/整数切片 Getter、有序 `Keys` 与来源查询。
- `ConfigFile` 的原文、格式、来源、变更订阅；properties、yaml/yml 支持 `AsMap`。
- Config Service 直连或 Meta `/services/config` 发现、轮转节点选择、`releaseKey`、`ip`、`label`、`dataCenter`、`messages` 与按 AppId 的 Access Key 签名。
- `FULL_SYNC` 和 `INCREMENTAL_SYNC`；增量结果在无全量基线、未知同步类型或非法变更类型时拒绝发布。
- 按 AppId 的 `/notifications/v2` 长轮询、Context/`Close` 取消、退避和通知触发刷新。
- 容灾链：Remote → 原子本地 JSON 缓存 → 可选 `ConfigMapStore`；Local Mode 只读取本地缓存。读取兼容旧 agollo JSON 缓存。
- 有界监听队列、溢出合并到最新事件、panic 隔离、精确 key/prefix/正则监听，以及无第三方依赖的 `Monitor` 快照。
- `Load` 显式预加载必需 namespace，替代隐式 `MustStart`；`RequestSigner` 和 `ConfigServiceSelector` 提供实例级认证与节点选择扩展。

旧测试中的运行时代码 patch 已改为实例级同步函数注入，移除了 `gomonkey` 依赖；这消除了 Go 新版本下的测试顺序不稳定性。

## 用例设计与覆盖

| 用例 | 验证点 | 自动化测试 |
| --- | --- | --- |
| 公开 API 契约 | 仅使用 `/v6` 导出标识符编译并运行 README 入口、多 AppId 与 ConfigFile | `TestPublicApolloClientAPI` |
| v6 Getter 与扩展点 | slice Getter、正则订阅、实例级 signer/selector | `TestApolloClientV6GetterAndExtensionContracts` |
| 启动预加载 | `Load` 依次加载所有必需 namespace | `TestApolloClientLoadEagerlyLoadsNamespaces` |
| ClientOptions 校验 | 服务地址、离线模式、退避、监听队列等非法参数快速失败 | `TestClientOptionsValidation` |
| 基础加载和鉴权 | Config URL、ip/DC/label、签名、类型化读取、来源、指标 | `TestApolloClientLoadsConfigAndAppliesProtocolParameters` |
| YAML ConfigFile | `.yaml` namespace、原文、扁平化 Map、原文变更事件 | `TestApolloClientConfigFileYAMLAndRawListener` |
| 多 AppId + 增量 | Secret 隔离、`messages`/`releaseKey` 回传、增加/修改/删除、key 前缀过滤 | `TestApolloClientMultiAppIDAndIncrementalSync` |
| 安全失败 | 无基线的增量同步不能覆盖内存配置 | `TestApolloClientRejectsIncrementalConfigWithoutBaseline` |
| 失败后的恢复 | 首次加载失败后可成功重试；成功后不再返回过期错误 | `TestApolloClientRetriesFailedInitialLoad` |
| 304 通知确认 | 保持内容和事件不变，同时推进 `notificationId` | `TestApolloClientAcknowledges304NotificationWithoutChangeEvent` |
| 刷新故障 | 已有远端快照时拒绝以旧缓存回滚 | `TestApolloClientRefreshFailureKeepsLastKnownGoodSnapshot` |
| Close 边界 | Close 后不得新增配置状态或 Config/ConfigFile 订阅 | `TestApolloClientCloseRejectsNewStateAndSubscriptions` |
| Cluster 隔离 | 拒绝加载不同 cluster 的缓存文件 | `TestDecodeDiskSnapshotRejectsDifferentCluster` |
| 本地缓存容灾 | 远端写入、原子缓存、Local Mode 离线恢复、来源标识 | `TestApolloClientFallsBackToAtomicLocalCache` |
| ConfigMap 容灾 | 远端异步落库、远端失败降级，以及无缓存目录的 Offline 加载 | `TestApolloClientPersistsAndFallsBackToConfigMapStore`、`TestApolloClientOfflineLoadsConfigMapWithoutCacheDirectory` |
| Meta 发现 | `/services/config` AppId/IP 参数和后续加载 | `TestApolloClientDiscoversConfigServiceFromMetaServer` |
| 长轮询 | 通知载荷 DC/IP、通知后刷新配置 | `TestApolloClientLongPollBuildsDataCenterAndRefreshes` |
| 生命周期 | `Close` 及时取消阻塞中的长轮询 | `TestApolloClientCloseCancelsLongPoll` |
| 既有能力回归 | 根包及全部子包 | `go test ./... -count=1` |
| 并发安全 | 新客户端并发访问、监听和 Close | `go test -race . -run '^TestApolloClient' -count=1` |

## 验证结果

以下命令已在本分支执行通过：

```text
go test ./... -count=1
go test -race . -run '^TestApolloClient|^TestPublicApolloClientAPI|^TestClientOptionsValidation' -count=1
go test ./protocol/http ./storage ./utils ./utils/parse/... -count=1
```

`go vet ./...` 仍会报告旧包中复制 `sync.Map`/`sync.Once` 的告警（`env`、`storage`、`component/serverlist`）；本次新增文件没有对应告警。它们应在后续 M0/M3 单独清理，不能据此宣称发布门禁已全部达成。

## 尚未完成的发布级工作

- Kubernetes `client-go` ConfigMap adapter（包括 RBAC、resourceVersion 冲突重试）和 Prometheus/OpenTelemetry exporter。
- 可复用 Mock Apollo TestKit、Java/Go 共享协议 fixture、跨 Apollo Server 版本的兼容矩阵。
- 多轮稳定性、全仓 `-race`、资源泄漏、性能和安全门禁，以及生产灰度和回滚演练。

这些工作及验收条件详见[完整迁移方案](agollo-refactor-java-client-migration-plan.md)。
