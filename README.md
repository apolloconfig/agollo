Agollo - Go Client for Apollo
================

[![Build Status](https://travis-ci.org/zouyx/agollo.svg?branch=master)](https://travis-ci.org/zouyx/agollo)


Apollo的Golang客户端。

Installation
------------

如果还没有安装Go开发环境，请参考以下文档[Getting Started](http://golang.org/doc/install.html) ，安装完成后，请执行以下命令：

```
//以下是正常逻辑使用
gopm get github.com/cihub/seelog -v -g
//以下是测试用例使用
gopm  get github.com/fatih/structs -v -g
gopm  get github.com/ajg/form -v -g
gopm  get github.com/gavv/monotime -v -g
gopm  get github.com/google/go-querystring/query -v -g
gopm  get github.com/imkira/go-interpol -v -g
gopm  get golang.org/x/net/publicsuffix -v -g
gopm  get github.com/moul/http2curl -v -g
gopm  get github.com/stretchr/testify -v -g
gopm  get github.com/valyala/fasthttp -v -g
gopm get "github.com/xeipuuv/gojsonschema" -v -g
gopm get "github.com/yalp/jsonpath" -v -g
gopm get "github.com/yudai/gojsondiff" -v -g
gopm get "github.com/yudai/gojsondiff/formatter" -v -g

```

*请注意*: 最好使用Go 1.8进行开发

# Features
* 实时同步配置
* 灰度配置

# Usage
  [使用指南](https://github.com/zouyx/agollo/wiki/使用指南)

# To Do
* 客户端容灾

# Contribution
  * Source Code: https://github.com/zouyx/agollo/
  * Issue Tracker: https://github.com/zouyx/agollo/issues
  
# License
The project is licensed under the [Apache 2 license](https://github.com/zouyx/agollo/blob/master/LICENSE).

# Reference
Apollo : https://github.com/ctripcorp/apollo
