package option

import (
	"reflect"
	"time"
)

const (
	LevelDisable = 0 //禁止
	LevelModel   = 1 //只缓存模型
	LevelSearch  = 2 //查询缓存
	MinExpires   = 30
	MaxExpires   = 3600
)

var (
	optMap map[reflect.Type]*ModelOption
)

// Opt .
type Opt struct {
	Expires         int  //默认120秒，30-900
	Level           int  //默认LevelSearch，LevelDisable:关闭，LevelModel:模型缓存， LevelSearch:查询缓存
	AsyncWrite      bool //默认false， insert update delete 成功后是否异步更新缓存
	PenetrationSafe bool //默认false, 开启防穿透。
}

// DefaultOption .
type DefaultOption struct {
	Opt
	redis struct {
		addr     string
		password string
		db       int
		option   *RedisOption
	}
}

type ModelOption struct {
	Opt
}

func (defOpt *DefaultOption) Init() {
	if defOpt.Level == 0 {
		defOpt.Level = LevelSearch
	}
	if defOpt.Expires == 0 {
		defOpt.Expires = 120
	}
	if defOpt.Expires < MinExpires {
		panic("minExpires 30")
	}
	if defOpt.Expires > MaxExpires {
		panic("maxExpires 900")
	}
}

// RedisOption .
type RedisOption struct {
	Addr string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration
}
