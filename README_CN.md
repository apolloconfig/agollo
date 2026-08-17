Agollo - Go Client for Apollo
================
中文 | [English](/README.md)

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![Build Status](https://github.com/apolloconfig/agollo/actions/workflows/go.yml/badge.svg)](https://github.com/apolloconfig/agollo/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/apolloconfig/agollo)](https://goreportcard.com/report/github.com/apolloconfig/agollo)
[![codebeat badge](https://codebeat.co/badges/bc2009d6-84f1-4f11-803e-fc571a12a1c0)](https://codebeat.co/projects/github-com-apolloconfig-agollo-master)
[![Coverage Status](https://coveralls.io/repos/github/apolloconfig/agollo/badge.svg?branch=master)](https://coveralls.io/github/apolloconfig/agollo?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](http://godoc.org/github.com/apolloconfig/agollo?status.svg)](http://godoc.org/github.com/apolloconfig/agollo)
[![GitHub release](https://img.shields.io/github/release/apolloconfig/agollo.svg)](https://github.com/apolloconfig/apolloconfig/releases)
[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)

方便Golang接入配置中心框架 [Apollo](https://github.com/ctripcorp/apollo) 所开发的Golang版本客户端。

# Features

* 支持多 IP、AppID、namespace
* 实时同步配置
* 灰度配置
* 延迟加载（运行时）namespace
* 客户端，配置文件容灾
* 自定义日志，缓存组件
* 支持配置访问秘钥
* 实例级 `ApolloClient`，支持多 AppId 隔离、类型化读取、ConfigFile 和 key/prefix 监听
* 支持全量/增量同步、长轮询及 Remote → 本地缓存 → 可选 ConfigMap 容灾

# Usage

## 快速入门

### 导入 agollo

```
go get github.com/apolloconfig/agollo/v6@latest
```

### 启动 agollo

新项目使用实例级 `ApolloClient`，避免进程全局状态并显式管理客户端生命周期：

```go
client, err := agollo.NewClient(context.Background(), agollo.ClientOptions{
	AppID:      "orders",
	Cluster:    "default",
	MetaServer: "http://apollo-meta:8080",
	CacheDir:   "/var/lib/orders/apollo",
})
if err != nil {
	panic(err)
}
defer client.Close()

cfg, err := client.Config(context.Background(), "application")
if err != nil {
	panic(err)
}
port := cfg.Int("server.port", 8080)
```

新版通过 `ConfigForApp` 支持单 Client 多 AppId，通过 `ConfigFile` 读取 YAML、JSON、XML、TXT 等 namespace 原文。需要启动即失败的服务可调用 `Load` 预加载必需 namespace。v5 到 v6 的字段映射、Getter、监听器和扩展点替代方式见[迁移指南](docs/migration-to-apollo-client.md)。

## 更多用法

***使用Demo*** ：[agollo_demo](https://github.com/zouyx/agollo_demo)

***其他语言*** ： [agollo-agent](https://github.com/zouyx/agollo-agent.git) 做本地agent接入，如：PHP

欢迎查阅 [Wiki](https://github.com/apolloconfig/agollo/wiki) 或者 [godoc](http://godoc.org/github.com/zouyx/agollo) 获取更多有用的信息

如果你觉得该工具还不错或者有问题，一定要让我知道，可以发邮件或者[留言](https://github.com/apolloconfig/agollo/issues)。

# User

* [使用者名单](https://github.com/apolloconfig/agollo/issues/20)

# Contribution

* Source Code: https://github.com/apolloconfig/agollo/
* Issue Tracker: https://github.com/apolloconfig/agollo/issues

# License

The project is licensed under the [Apache 2 license](https://github.com/apolloconfig/agollo/blob/master/LICENSE).

# Reference

Apollo: [https://github.com/apolloconfig/apollo](https://github.com/apolloconfig/apollo)
