# 从 agollo v5 迁移到 v6 ApolloClient

v6 使用 Go 的语义化导入版本，模块路径从 `/v5` 改为 `/v6`。本文面向正在使用 `Start`、`StartWithConfig`、旧 `Client` 和 `storage.Config` 的项目。v6 的公开文档和新功能只承诺 `ApolloClient` API；请在升级模块路径时完成本指南中的调用迁移。

```go
// v5
import "github.com/apolloconfig/agollo/v5"

// v6
import "github.com/apolloconfig/agollo/v6"
```

```sh
go get github.com/apolloconfig/agollo/v6@latest
```

## 为什么重新设计 API

新版入口采用以下形态：

```go
client, err := agollo.NewClient(ctx, agollo.ClientOptions{...})
```

`NewClient` 返回具体类型 `*ApolloClient`，而不是要求用户依赖一个包含全部方法的大接口。需要 mock 时，业务代码应在消费侧声明只包含所需方法的小接口。`ClientOptions` 集中展示全部启动配置，便于代码审查、配置映射和后续增加可选字段；运行中的客户端不会读取或修改传入结构。

读取方法采用 Go 风格命名：

| 用途 | 新 API |
| --- | --- |
| 默认 AppId 的 KV namespace | `client.Config(ctx, namespace)` |
| 指定 AppId 的 KV namespace | `client.ConfigForApp(ctx, appID, namespace)` |
| 默认 AppId 的原始文件 | `client.ConfigFile(ctx, namespace, format)` |
| 指定 AppId 的原始文件 | `client.ConfigFileForApp(ctx, appID, namespace, format)` |
| 客户端状态 | `client.Monitor().Snapshot()` |
| 释放资源 | `client.Close()` |

业务需要替身实现时，在消费侧定义最小接口：

```go
type ConfigProvider interface {
    Config(context.Context, string) (agollo.Config, error)
}
```

这样新增 `ApolloClient` 能力不会迫使业务 mock 同步增加无关方法。

## 最小迁移示例

旧代码：

```go
appConfig := &config.AppConfig{
    AppID:          "orders",
    Cluster:        "default",
    IP:             "http://apollo-meta:8080",
    NamespaceName:  "application",
    IsBackupConfig: true,
    BackupConfigPath: "/var/lib/orders/apollo",
    Secret:         "secret",
}

client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
    return appConfig, nil
})
if err != nil {
    return err
}
defer client.Close()

cache := client.GetConfigCache("application")
value, err := cache.Get("server.port")
```

新代码：

```go
client, err := agollo.NewClient(ctx, agollo.ClientOptions{
    AppID:          "orders",
    Cluster:        "default",
    MetaServer:     "http://apollo-meta:8080",
    CacheDir:       "/var/lib/orders/apollo",
    AccessKeySecret: "secret",
})
if err != nil {
    return err
}
defer client.Close()

cfg, err := client.Config(ctx, "application")
if err != nil {
    return err
}
port := cfg.Int("server.port", 8080)
value, exists := cfg.Lookup("server.port")
```

`Config` 在首次调用时加载 namespace，后续调用返回同一个实时配置视图。`Lookup` 用 `exists` 区分“配置不存在”和“配置值为空”；类型化 Getter 在不存在或转换失败时返回调用方提供的默认值。

需要与旧 `MustStart` 一样在启动阶段验证配置时，显式预加载：

```go
if err := client.Load(ctx, "application", "feature.properties"); err != nil {
    return fmt.Errorf("load required Apollo configuration: %w", err)
}
```

## AppConfig 字段映射

| 旧 `config.AppConfig` | 新 `ClientOptions` | 迁移说明 |
| --- | --- | --- |
| `AppID` | `AppID` | 作为默认 AppId；也可留空并只使用 `ConfigForApp` |
| `Cluster` | `Cluster` | 空值默认 `default` |
| `IP` | `MetaServer` | 旧配置通常指向 Meta Server；直连 Config Service 时使用 `ConfigServices` |
| `NamespaceName` | 无启动字段 | 在需要时调用 `Config`/`ConfigFile`；多个 namespace 不再拼接字符串 |
| `IsBackupConfig` | 是否设置 `CacheDir` | 设置后启用原子本地缓存及首次加载降级 |
| `BackupConfigPath` | `CacheDir` | 新缓存按 AppId、cluster、namespace、格式隔离，并兼容读取旧 JSON |
| `Secret` | `AccessKeySecret` | 多 AppId 使用 `AccessKeySecrets[appID]` 覆盖 |
| `Label` | `Label` | 灰度发布语义不变 |
| `SyncServerTimeout` | Context / `HTTPClient` | 单次读取使用调用 Context；自定义 HTTP Client 的总超时需允许 90 秒长轮询 |
| `MustStart` | 启动阶段显式调用 `Config` 并处理错误 | 新构造函数不执行网络请求，是否阻止进程启动由业务决定 |

`DataCenter` 和 `ClientIP` 是新版显式字段，会同时用于配置拉取和长轮询。`ConfigServices` 非空时跳过 Meta 发现并按顺序轮转地址。

## ClientOptions 默认值和高级字段

| 字段 | 零值行为 | 说明 |
| --- | --- | --- |
| `AppID` | 无默认 AppId | 只使用 `ConfigForApp` 时允许为空 |
| `Cluster` | `default` | 当前 Client 内所有 AppId 共用该 cluster |
| `ConfigServices` | 未配置 | 直连地址；非空时优先于 `MetaServer` |
| `MetaServer` | 未配置 | 非 Offline 模式下两类服务地址至少配置一种 |
| `AccessKeySecret` | 不签名 | 默认 AppId Secret |
| `AccessKeySecrets` | 空 map | 按 AppId 覆盖 Secret，空字符串可显式关闭该 AppId 签名 |
| `HTTPClient` | 标准 `http.Client` | Transport 必须支持并发；总超时应允许长轮询 |
| `CacheDir` | 不启用文件缓存 | 目录权限创建为 `0750`，缓存文件权限 `0640` |
| `ConfigMapStore` | 不启用 | Remote 和文件均失败时用于首次加载降级 |
| `Offline` | `false` | `true` 时要求 `CacheDir` 或 `ConfigMapStore`，且不启动发现和长轮询 |
| `DisableLongPolling` | `false` | CLI/批处理或确定性测试可关闭 |
| `LongPollInitialDelay` | `0` | 首次成功加载后延迟启动长轮询；不能为负 |
| `RetryBackoffMin/Max` | `1s` / `2m` | 有界指数退避范围 |
| `ListenerQueueSize` | `32` | 必须非负；`0` 使用默认值 |
| `RequestSigner` | `nil` | 实例级请求签名器；非空时替代内置 Access Key 签名 |
| `ConfigServiceSelector` | 轮转 | 实例级 Config Service 选择策略；必须返回当前服务列表中的地址 |

## 常见调用迁移

### 配置读取

| 旧调用 | 新调用 |
| --- | --- |
| `client.GetConfigCache(ns).Get(key)` | `cfg, err := client.Config(ctx, ns)`；`cfg.Lookup(key)` |
| `client.GetStringValue(key, def)` | `cfg.String(key, def)` |
| `client.GetIntValue(key, def)` | `cfg.Int(key, def)` |
| `client.GetFloatValue(key, def)` | `cfg.Float64(key, def)` |
| `client.GetBoolValue(key, def)` | `cfg.Bool(key, def)` |
| `client.GetStringSliceValue(key, def)` | `cfg.StringSlice(key, def)` |
| `client.GetIntSliceValue(key, def)` | `cfg.IntSlice(key, def)` |
| 遍历自定义 Cache | `cfg.Keys()` 或 `cfg.Snapshot()` |

旧 Client 的快捷 Getter 固定读取默认 namespace；新 API 要求先取得明确的 namespace，减少跨 namespace 误读。

### 变更监听

旧监听器挂在整个 Client 上：

```go
client.AddChangeListener(listener)
```

新版监听器属于一个 namespace，并返回取消函数：

```go
cfg, err := client.Config(ctx, "application")
if err != nil {
    return err
}

cancel := cfg.Subscribe(func(event agollo.ConfigChangeEvent) {
    // event.Changes 包含同一原子快照中的完整变更集
},
    agollo.WithInterestedKeys("server.port"),
    agollo.WithInterestedKeyPrefixes("db."),
	    agollo.WithInterestedKeyRegexps(regexp.MustCompile(`^feature\\.`)),
)
defer cancel()
```

监听器使用有界队列。慢监听器不会无限创建 goroutine；队列满时合并为最新事件，并增加 `MonitorSnapshot.ListenerDrops`。

### ConfigFile

Java Client 风格的原始文件能力通过独立视图暴露：

```go
file, err := client.ConfigFile(ctx, "application", agollo.ConfigFileFormatYAML)
if err != nil {
    return err
}

raw := file.Content()
cancel := file.Subscribe(func(event agollo.ConfigFileChangeEvent) {
    // 使用 event.NewContent 原文
})
defer cancel()
```

支持 `properties`、`xml`、`json`、`yml`、`yaml`、`txt`。原文 API 不承诺将 XML/JSON/TXT 转成 KV；properties 和 yaml/yml 可通过 `AsMap` 读取。

### 单进程多 AppId

```go
client, err := agollo.NewClient(ctx, agollo.ClientOptions{
    ConfigServices: []string{"http://apollo-config:8080"},
    AccessKeySecrets: map[string]string{
        "orders":   "orders-secret",
        "payments": "payments-secret",
    },
})

orders, err := client.ConfigForApp(ctx, "orders", "application")
payments, err := client.ConfigForApp(ctx, "payments", "application")
```

AppId 是配置身份和签名选择的一部分，不会共用通知 ID、release key 或本地缓存文件。

## 错误处理和生命周期变化

- `NewClient` 只校验参数、初始化实例，不访问网络。建议在应用启动阶段主动调用所有必需 namespace 的 `Config`，用返回错误实现旧 `MustStart` 语义。
- 首次远端加载失败时，顺序尝试本地缓存和可选 `ConfigMapStore`。已有内存快照后的刷新失败不会回滚到更旧的容灾副本。
- 调用 Context 控制首次加载；传给 `NewClient` 的 Context 控制客户端整体生命周期。仍应调用 `Close`，等待监听器、长轮询和异步持久化任务退出。
- `Close` 幂等；Close 后创建新 namespace 或订阅会失败或成为空操作。
- 新 API 不使用全局扩展 setter。自定义网络行为通过 `HTTPClient`，认证通过 `RequestSigner`，实例级节点选择通过 `ConfigServiceSelector`，Kubernetes 容灾通过 `ConfigMapStore` 注入。

## 扩展点替代

| v5 全局扩展 | v6 实例级替代 |
| --- | --- |
| `SetSignature` | `ClientOptions.RequestSigner`；默认仍支持 `AccessKeySecret` |
| `SetLoadBalance` | `ClientOptions.ConfigServiceSelector`；默认轮转 |
| `SetLogger` | 由调用方的 `HTTPClient` Transport、监控采集和应用日志统一处理；核心不修改进程全局 logger |
| `SetCache` / `SetBackupFileHandler` | `CacheDir` 的原子本地快照与 `ConfigMapStore`；不再暴露可变 Cache |
| `AddFormatParser` | `ConfigFile` 保留原文；调用方按文件格式选择解析器 |

## 建议的渐进迁移顺序

1. 在独立分支将模块 import 从 `/v5` 替换为 `/v6`，先完成编译。
2. 迁移只读 Getter，按 namespace 建立 `Config` 变量，并用 `Load` 替代 `MustStart`。
3. 迁移监听器，并在组件退出时调用订阅取消函数；正则筛选使用 `WithInterestedKeyRegexps`。
4. 如使用非 properties namespace，迁移到 `ConfigFile` 并核对原文语义。
5. 以 `RequestSigner`、`ConfigServiceSelector` 和 `ConfigMapStore` 替换进程级扩展。
6. 验证远端不可用时的首次加载、本地缓存目录权限和关闭流程后，删除业务中的旧 `StartWithConfig` 与旧缓存依赖。

迁移期间不要让同一业务逻辑同时消费新旧监听事件，否则一次 Apollo 发布可能被处理两次。建议在灰度环境比较新旧读取结果和事件序列，再逐个关闭旧调用路径。
