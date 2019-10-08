# gcache
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gcache/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree) 

###### 
###### Plug and play caching middleware for gorm.

## Overview
- Plug and Play
- Cache aside
- Use `Redis` as data source
- Prevent hotspot data set is invalid
- Prevent cache penetration 

#### Installation
```sh
$ go get -u github.com/8treenet/gcache
```
#### Quick start
```go
import (
    "github.com/8treenet/gcache"
    "github.com/jinzhu/gorm"
    "github.com/8treenet/gcache/option""
)

func init() {
    //create gorm.DB
    db, _ = gorm.Open("mysql", "")

    opt := option.DefaultOption{}
    opt.Expires = 300                //cache expire time, the default value is 60 second and you can change it from 30s to 900s.
    opt.Level = option.LevelSearch   //cache level，the default is `LevelSearch`。`LevelDisable`:close the cache, `LevelModel`:model cache， `LevelSearch`:search cache.
    opt.AsyncWrite = false           //update cache asynchronously, the default is `false`。 insert update delete update cache or not if your row has been inserted successfully. 
    opt.PenetrationSafe = false 	 //open the `Prevent hotspot data set is invalid` mode or not, the default value is false.
    
    //add this middleware to gorm.DB
    gcache.AttachDB(db, &opt, &option.RedisOption{Addr:"localhost:6379"})
}
```

#### Notice
- `Group` is not supported.
- `Having` is not supported.
- Query staement and query parameters is separated.


#### Example
```shell script
    #read `example_test.go` to learn more:
    src/github.com/8treenet/gcache/example/example_test.go
```
