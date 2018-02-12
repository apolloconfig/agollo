Agollo - Go Client for Apollo
================

[![Build Status](https://travis-ci.org/zouyx/agollo.svg?branch=master)](https://travis-ci.org/zouyx/agollo)
[![codebeat badge](https://codebeat.co/badges/bc2009d6-84f1-4f11-803e-fc571a12a1c0)](https://codebeat.co/projects/github-com-zouyx-agollo-master)
[![Coverage Status](https://coveralls.io/repos/github/zouyx/agollo/badge.svg?branch=master)](https://coveralls.io/github/zouyx/agollo?branch=master)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](http://godoc.org/github.com/zouyx/agollo?status.svg)](http://godoc.org/github.com/zouyx/agollo)
[![GitHub release](https://img.shields.io/github/release/zouyx/agollo.svg)](https://github.com/zouyx/agollo/releases)

方便Golang接入配置中心框架 [Apollo](https://github.com/ctripcorp/apollo) 所开发的Golang版本客户端。

Installation
------------

如果还没有安装Go开发环境，请参考以下文档[Getting Started](http://golang.org/doc/install.html) ，安装完成后，请执行以下命令：

``` shell
gopm get github.com/cihub/seelog -v -g
gopm get github.com/coocood/freecache -v -g
gopm get github.com/zouyx/agollo -v -g
```

或者

``` shell
go get -u github.com/cihub/seelog
go get -u github.com/coocood/freecache
go get -u github.com/zouyx/agollo
```


*请注意*: 最好使用Go 1.8进行开发

# Features
* 实时同步配置
* 灰度配置
* 客户端容灾

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
	 go agollo.StartWithLogger(loggerInterface)
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
