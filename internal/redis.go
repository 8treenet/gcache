package internal

import (
	"time"

	"github.com/8treenet/gcache/option"
	"github.com/go-redis/redis"
)

// RedisClient .
type RedisClient interface {
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(script string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	FlushDB() *redis.StatusCmd
	HGetAll(key string) *redis.StringStringMapCmd
	HDel(key string, fields ...string) *redis.IntCmd
	HScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd
	HGet(key, field string) *redis.StringCmd
	MGet(keys ...string) *redis.SliceCmd
	HSet(key, field string, value interface{}) *redis.BoolCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
}

func newRedisClient(option *option.RedisOption) (result RedisClient) {
	client := redis.NewClient(&redis.Options{
		Addr:               option.Addr,
		Password:           option.Password,
		DB:                 option.DB,
		MaxRetries:         option.MaxRetries,
		PoolSize:           option.PoolSize,
		ReadTimeout:        option.ReadTimeout,
		WriteTimeout:       option.WriteTimeout,
		MinIdleConns:       option.MinIdleConns,
		MaxConnAge:         option.MaxConnAge,
		IdleTimeout:        option.IdleTimeout,
		IdleCheckFrequency: option.IdleCheckFrequency,
		PoolTimeout:        option.PoolTimeout,
	})

	if perr := client.Ping().Err(); perr != nil {
		panic(perr)
	}

	result = client
	return
}
