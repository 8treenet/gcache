# gcache
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gcache/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree) 

###### gcache是gorm的中间件，插入后gorm即刻拥有缓存。

## Overview
- 即插即用
- 旁路缓存
- 数据源使用 Redis
- 防击穿
- 防穿透

#### 安装
```sh
$ go get github.com/8treenet/gcache
```
#### 快速使用
```go
import (
    "github.com/8treenet/gcache"
    "github.com/jinzhu/gorm"
    "github.com/8treenet/gcache/option""
)

func init() {
    //创建 gorm.DB
    db, _ = gorm.Open("mysql", "")

    opt := option.DefaultOption{}
    opt.Expires = 300                //缓存时间，默认120秒。范围 30-3600
    opt.Level = option.LevelSearch   //缓存级别，默认LevelSearch。LevelDisable:关闭缓存，LevelModel:模型缓存， LevelSearch:查询缓存
    opt.AsyncWrite = false           //异步缓存更新, 默认false。 insert update delete 成功后是否异步更新缓存。 ps: affected如果未0，不触发更新。
    opt.PenetrationSafe = false 	 //开启防穿透, 默认false。 ps:防击穿强制全局开启。
    
    //缓存中间件附加到gorm.DB
    gcache.AttachDB(db, &opt, &option.RedisOption{Addr:"localhost:6379"})
}
```

#### 约定
- 模型必须指明主键 `gorm:"primary_key"`
- 不支持 Group
- 不支持 Having
- 查询条件和查询参数分离


#### [Example](https://github.com/8treenet/gcache/blob/master/example/example_test.go)
```shell script
    #查看 example_test.go 了解更多。
    more src/github.com/8treenet/gcache/example/example_test.go
```
