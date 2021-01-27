package internal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TIZX/gcache/option"
	"github.com/go-redis/redis"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type TestEmail struct {
	gorm.Model
	TypeID     int
	Subscribed bool
	TestUserID int
}

var modelValue string = `
{\"PK\":\"18\",\"Model\":{\"ID\":18,\"CreatedAt\":\"2019-09-15T16:30:16+08:00\",\"Timeout\":\"2019-09-15T16:30:16+08:00\",\"DeletedAt\":null,\"TypeID\":18,\"Subscribed\":false,\"TestUserID\":18}}
`

func TestInitRefreshData(t *testing.T) {
	cp := gettestcachePlugin()
	cp.FlushDB()
	cp.handle.redisClient.Set("test_emails_model_18", modelValue, 300*time.Second).Err()
	var js JsonSearch
	js.Primarys = append(js.Primarys, 18)
	js.Timeout = time.Now().Unix() + 30
	buffer, _ := json.Marshal(js)
	cp.handle.redisClient.HSet("test_emails_search_&((type_id>=$)", "18_LIMIT_1", buffer)
	cp.handle.redisClient.Expire("test_emails_search_&((type_id>=$)", 300*time.Second)

	cp.handle.redisClient.HSet("test_emails_affect_type_id", "test_emails_search_&((type_id>=$)", js.Timeout)
}

func TestRefresh(t *testing.T) {
	cp := gettestcachePlugin()
	cp.Debug()

	dh := cp.handle.NewDeleteHandle()
	dh.refresh("test_emails_search_&((type_id>=$)", true)
	dh.refresh("test_emails_affect_type_id", false)
}

func gettestcachePlugin() *plugin {
	addr := "root:123123@tcp(127.0.0.1:3306)/matrix?charset=utf8&parseTime=True&loc=Local"
	db, e := gorm.Open("mysql", addr)
	if e != nil {
		panic(e)
	}

	opt := option.DefaultOption{}
	//缓存插件 注入到Gorm。开启Debug，查看日志
	cachePlugin := InjectGorm(db, &opt, &option.RedisOption{Addr: "localhost:6379"})
	return cachePlugin
}

func TestLuaSetAffect(t *testing.T) {
	c := gettestcachePlugin()
	script := redis.NewScript(`
	for i=1,100000,1 do
		redis.call("HSET", KEYS[1], i, ARGV[1])
	end
	return true
`)
	t.Log(script.Run(c.handle.redisClient, []string{"www"}, time.Now().Unix()-100).Result())
}

func TestLuaAffectRefresh(t *testing.T) {
	c := gettestcachePlugin()
	script := redis.NewScript(`
	local all = redis.call("HGETALL", KEYS[1])
	local key = ""
	local timeout = tonumber(ARGV[1])
	for k,v in pairs(all) do
		if k % 2 == 1 then
			key = v
		else
			if tonumber(v) < timeout then
				redis.call("HDEL", KEYS[1], key)
			end
		end
	end
	return true
`)
	//redis.log(redis.LOG_NOTICE, key, v)
	t.Log(script.Run(c.handle.redisClient, []string{"www"}, time.Now().Unix()-50).Result())
}

func TestLuaSetSearch(t *testing.T) {
	c := gettestcachePlugin()
	script := redis.NewScript(`
	for i=1,100,1 do
		redis.call("HSET", KEYS[1], i, ARGV[1])
	end
`)
	var js JsonSearch
	js.Primarys = append(js.Primarys, 18)
	js.Timeout = time.Now().Unix() - 100
	buffer, _ := json.Marshal(js)
	t.Log(script.Run(c.handle.redisClient, []string{"com"}, string(buffer)).Result())
}

func TestLuaSearchRefresh(t *testing.T) {
	c := gettestcachePlugin()
	script := redis.NewScript(`
	local all = redis.call("HGETALL", KEYS[1])
	local key = ""
	local timeout = tonumber(ARGV[1])
	for k,v in pairs(all) do
		if k % 2 == 1 then
			key = v
		else
			local data = cjson.decode(v);
			if tonumber(data["Timeout"]) < timeout then
				redis.call("HDEL", KEYS[1], key)
			end
		end
	end
	return true
`)
	//redis.log(redis.LOG_NOTICE, key, v)
	t.Log(script.Run(c.handle.redisClient, []string{"com"}, time.Now().Unix()-150).Result())
}
