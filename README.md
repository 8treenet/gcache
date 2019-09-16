# gcache
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gcache/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree) [![QQ群](https://img.shields.io/:QQ%E7%BE%A4-602434016-blue.svg)](https://github.com/8treenet/jaguar) 

###### gcache是gorm的缓存插件，注入即可轻松使用。

## Overview
- 即插即用
- 旁路缓存
- 数据源使用 Redis
- 防击穿

#### 安装
```sh
$ go get -u github.com/8treenet/gcache
```
#### 快速使用
```go
import (
    "github.com/8treenet/gcache"
    "github.com/jinzhu/gorm"
    "github.com/8treenet/gcache/option""
)

func init() {
    gormdb, _ = gorm.Open("mysql", "")
    opt := option.DefaultOption{}
    opt.Expires = 300                //缓存时间，默认60秒。范围 30-900
    opt.Level = option.LevelSearch   //缓存级别，默认LevelSearch。LevelDisable:关闭缓存，LevelModel:模型缓存， LevelSearch:查询缓存
    opt.AsyncWrite = false           //异步缓存更新, 默认false。 insert update delete 成功后是否异步更新缓存
    opt.RedisAddr = "localhost:6379" //redis 地址
    opt.RedisPassword = ""           //redis 密码
    opt.RedisDB = 0                  //redis 库
    
    //缓存插件 注入到Gorm。
    gcache.InjectGorm(gormdb, &opt)
}
```

#### 约定
- 不支持 Gruop
- 不支持 Having
- 查询条件和查询参数分离


#### Example
```shell script
    #查看 example_test.go 了解更多。
    more src/github.com/8treenet/gcache/example/example_test.go
```