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

# Usage

## Quick Start

### Import agollo

```
go get -u github.com/apolloconfig/agollo/v5@latest
```

### Initialize agollo

```go
package main

import (
	"fmt"

	"github.com/apolloconfig/agollo/v5"
	"github.com/apolloconfig/agollo/v5/env/config"
)

func main() {
	c := &config.AppConfig{
		AppID:          "testApplication_yang",
		Cluster:        "dev",
		IP:             "http://localhost:8080",
		NamespaceName:  "dubbo",
		IsBackupConfig: true,
		Secret:         "6ce3ff7e96a24335a9634fe9abca6d51",
	}

	client, _ := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	fmt.Println("Apollo configuration initialized successfully")

	//Use your apollo key to test
	cache := client.GetConfigCache(c.NamespaceName)
	value, _ := cache.Get("key")
	fmt.Println(value)
}
```

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
