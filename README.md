Agollo - Go Client for Apollo
================
English | [中文](/README_CN.md)

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![Build Status](https://github.com/apolloconfig/agollo/actions/workflows/go.yml/badge.svg)](https://github.com/apolloconfig/agollo/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/apolloconfig/agollo)](https://goreportcard.com/report/github.com/apolloconfig/agollo)
[![codebeat badge](https://codebeat.co/badges/bc2009d6-84f1-4f11-803e-fc571a12a1c0)](https://codebeat.co/projects/github-com-apolloconfig-agollo-master)
[![Coverage Status](https://coveralls.io/repos/github/apolloconfig/agollo/badge.svg?branch=master)](https://coveralls.io/github/apolloconfig/agollo?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](http://godoc.org/github.com/apolloconfig/agollo?status.svg)](http://godoc.org/github.com/apolloconfig/agollo)
[![GitHub release](https://img.shields.io/github/release/apolloconfig/agollo.svg)](https://github.com/apolloconfig/apolloconfig/releases)
[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)

A Golang client for the configuration center framework [Apollo](https://github.com/apolloconfig/apollo).

# Features

* Support for multiple IPs, AppIDs, and namespaces
* Real-time configuration synchronization
* Gray release configuration
* Lazy loading (runtime) namespaces
* Client-side and configuration file fallback
* Customizable logger and cache components
* Support for configuration access keys
* Instance-scoped `ApolloClient` API with multi-AppID isolation, typed getters, `ConfigFile`, and key/prefix change subscriptions
* Config Service discovery, Access Key signing, incremental sync, long polling, and Remote → local cache → optional ConfigMap fallback

# Usage

## Quick Start

### Import agollo

```
go get github.com/apolloconfig/agollo/v6@latest
```

### Initialize agollo

New integrations should use the instance-scoped `ApolloClient` to avoid
process-global state and explicitly own the client lifecycle.

```go
package main

import (
	"context"
	"fmt"

	"github.com/apolloconfig/agollo/v6"
)

func main() {
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

	config, err := client.Config(context.Background(), "application")
	if err != nil {
		panic(err)
	}
	fmt.Println(config.Int("server.port", 8080))

	// Subscribe only to changes under the db. prefix.
	cancel := config.Subscribe(func(event agollo.ConfigChangeEvent) {
		fmt.Printf("configuration changed: %#v\n", event.Changes)
	}, agollo.WithInterestedKeyPrefixes("db."))
	defer cancel()
}
```

`ApolloClient` additionally supports multiple AppIDs through `ConfigForApp`,
raw namespace files through `ConfigFile`, and source inspection through
`Config.Source()`. It keeps the last successful in-memory snapshot on a refresh
failure; local cache and `ConfigMapStore` are used only to establish the first
usable snapshot. See the [v5 to v6 migration guide](docs/migration-to-apollo-client.md),
[migration plan](docs/agollo-refactor-java-client-migration-plan.md),
and [implementation/test matrix](docs/agollo-java-client-parity-implementation.md)
for the v6 options, migration scope, and validation coverage. `Load` can
eagerly load required namespaces when the application needs fail-fast startup.

## More Examples

***Demo Project***: [agollo_demo](https://github.com/zouyx/agollo_demo)

***Other Languages:***: Use [agollo-agent](https://github.com/zouyx/agollo-agent.git) as a local agent for languages like PHP.

Check out our [Wiki](https://github.com/apolloconfig/agollo/wiki) or [godoc](http://godoc.org/github.com/zouyx/agollo) for more information.

If you find this tool useful or encounter any issues, please let me know via email or by [creating an issue](https://github.com/apolloconfig/agollo/issues)。

# User

* [User List](https://github.com/apolloconfig/agollo/issues/20)

# Contribution

* Source Code: https://github.com/apolloconfig/agollo/
* Issue Tracker: https://github.com/apolloconfig/agollo/issues

# License

The project is licensed under the [Apache 2 license](https://github.com/apolloconfig/agollo/blob/master/LICENSE).

# Reference

Apollo: [https://github.com/apolloconfig/apollo](https://github.com/apolloconfig/apollo)
