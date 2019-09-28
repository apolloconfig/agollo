Agollo - Go Client for Apollo
================

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![Build Status](https://travis-ci.org/zouyx/agollo.svg?branch=master)](https://travis-ci.org/zouyx/agollo)
[![Go Report Card](https://goreportcard.com/badge/github.com/zouyx/agollo)](https://goreportcard.com/report/github.com/zouyx/agollo)
[![codebeat badge](https://codebeat.co/badges/bc2009d6-84f1-4f11-803e-fc571a12a1c0)](https://codebeat.co/projects/github-com-zouyx-agollo-master)
[![Coverage Status](https://coveralls.io/repos/github/zouyx/agollo/badge.svg?branch=master)](https://coveralls.io/github/zouyx/agollo?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](http://godoc.org/github.com/zouyx/agollo?status.svg)](http://godoc.org/github.com/zouyx/agollo)
[![GitHub release](https://img.shields.io/github/release/zouyx/agollo.svg)](https://github.com/zouyx/agollo/releases)
[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)

方便Golang接入配置中心框架 [Apollo](https://github.com/ctripcorp/apollo) 所开发的Golang版本客户端。

***其他语言*** ： 可使用 [agollo-agent](https://github.com/zouyx/agollo-agent.git) 做本地agent接入。

Installation
------------

如果还没有安装Go开发环境，请参考以下文档[Getting Started](http://golang.org/doc/install.html) ，安装完成后，请执行以下命令：

``` shell
gopm get github.com/coocood/freecache -v -g
```

或者

``` shell
go get -u github.com/coocood/freecache
```

或者

```shell
chmod u+x ./depend.sh
./depend.sh
```


*请注意*: 最好使用Go 1.8进行开发

# Features
* 实时同步配置
* 灰度配置
* 客户端容灾
* 配置文件容灾 (v1.6.0+)
* 自定义日志组件 （因开源协议原因，从v1.8.0不再存在默认日志组件。如还需使用seelog，请参考：[如何使用seelog](https://github.com/zouyx/agollo/wiki/%E4%BD%BF%E7%94%A8seelog%E6%97%A5%E5%BF%97%E7%BB%84%E4%BB%B6)）

# Usage

- 异步启动agollo

场景：启动程序不依赖加载Apollo的配置。

``` go
func main() {
	 go agollo.Start()
}
```

- 同步启动agollo（v1.2.0+）

场景：启动程序依赖加载Apollo的配置。例：初始化程序基础配置。

``` go
func main() {
	 agollo.Start()
}
```

- 启动agollo - 自定义logger控件（感谢 @Adol1111 提供）

``` go
func main() {
	 agollo.StartWithLogger(loggerInterface)
}
```

- 启动agollo - 自定义cache控件 (v1.7.0+)

``` go
func main() {
	 agollo.StartWithCache(cacheInterface)
}
```

- 启动agollo - 自定义各种控件 (v1.8.0+)

```go
func main() {
    agollo.SetLogger(loggerInterface)
	agollo.SetCache(cacheInterface)
	agollo.Start()
}
```

- 监听变更事件（阻塞）

``` go
func main() {
	event := agollo.ListenChangeEvent()
	changeEvent := <-event
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))
}
```

- 获取Apollo的配置
  - String
  
  ```
  agollo.GetStringValue(Key,DefaultValue)
  ```
  - Int
  
  ```
  agollo.GetIntValue(Key,DefaultValue)
  ```

  - Float
  
  ```
  agollo.GetFloatValue(Key,DefaultValue)
  ```

  - Bool
  
  ```
  agollo.GetBoolValue(Key,DefaultValue)
  ```
  
  后续可支持更多类型
 
  欢迎查阅 [Wiki](https://github.com/zouyx/agollo/wiki) 或者 [godoc](http://godoc.org/github.com/zouyx/agollo) 获取更多有用的信息
  
  如果你觉得该工具还不错或者有问题，一定要让我知道，可以发邮件或者[留言](https://github.com/zouyx/agollo/issues)。

# User

* [使用者名单](https://github.com/zouyx/agollo/issues/20)

# Contribution
  * Source Code: https://github.com/zouyx/agollo/
  * Issue Tracker: https://github.com/zouyx/agollo/issues
  
# License
The project is licensed under the [Apache 2 license](https://github.com/zouyx/agollo/blob/master/LICENSE).

# Reference
Apollo : https://github.com/ctripcorp/apollo
