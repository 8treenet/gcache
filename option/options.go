package option

import (
	"reflect"
)

const (
	LevelDisable      = 0                           //禁止
	LevelModel        = 1                           //只缓存模型
	LevelSearch       = 2                           //查询缓存
	MinExpires        = 30
	MaxExpires        = 900
)

var (
	optMap map[reflect.Type]*ModelOption
)

type Opt struct {
	Expires    int  //默认60秒，30-900
	Level      int  //默认LevelSearch，LevelDisable:关闭，LevelModel:模型缓存， LevelSearch:查询缓存
	AsyncWrite bool //默认false， insert update delete 成功后是否异步更新缓存
}

type DefaultOption struct {
	Opt
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

type ModelOption struct {
	Opt
}

func (defOpt *DefaultOption) Init() {
	if defOpt.Level == 0 {
		defOpt.Level = LevelSearch
	}
	if defOpt.Expires == 0 {
		defOpt.Expires = 300
	}
	if defOpt.Expires < MinExpires {
		panic("minExpires 30")
	}
	if defOpt.Expires > MaxExpires {
		panic("maxExpires 900")
	}
}
