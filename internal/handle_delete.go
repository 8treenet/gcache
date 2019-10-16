package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

func newDeleteHandle(handle *Handle) *deleteHandle {
	dh := &deleteHandle{handle: handle}
	dh.loadLua()
	return dh
}

type deleteHandle struct {
	handle           *Handle
	delSha           string
	delIndexSha      string
	refreshAffectSha string
	refreshSearchSha string
}

func (dh *deleteHandle) flushDB() error {
	return dh.handle.redisClient.FlushDB().Err()
}

func (dh *deleteHandle) delModle(table string, primarys ...interface{}) error {
	if len(primarys) == 0 {
		return errors.New("primarys empty")
	}

	keys := dh.handle.JoinModelKeys(table, primarys...)
	e := dh.handle.redisClient.Del(keys...).Err()
	dh.handle.Debug("Delete model cache Key :", keys, "error :", e)
	return e
}

//func (dh *deleteHandle) DeleteSearch(table string, sfs []*gorm.StructField) error {
//for index := 0; index < len(sfs); index++ {
//	var delSearch []string
//	var err error
//	key := dh.handle.JoinAffectKey(table, sfs[index].DBName)
//
//	delSearch, err = dh.handle.redisClient.HKeys(key).Result()
//	if err != nil {
//		dh.handle.Debug(err)
//		continue
//	}
//	if len(delSearch) == 0 {
//		continue
//	}
//
//	err = dh.handle.redisClient.Del(delSearch...).Err()
//	dh.handle.Debug("Delete search cache Key :", delSearch, "error :", err)
//	if err != nil {
//		dh.handle.Debug(err)
//	}
//}
//return nil
//}

func (dh *deleteHandle) DeleteSearchByScope(scope *easyScope) error {
	table := scope.Table
	sfs := scope.GetStructFields()
	return dh.DeleteSearch(table, sfs, scope.indexKeys)
}

func (dh *deleteHandle) refresh(key string, search bool) {
	if search {
		_, e := dh.handle.redisClient.EvalSha(dh.refreshSearchSha, []string{key}, time.Now().Unix()).Result()
		dh.handle.Debug("Refresh search script execution, keys :", key, "error :", e)
		return
	}

	_, e := dh.handle.redisClient.EvalSha(dh.refreshAffectSha, []string{key}, time.Now().Unix()).Result()
	dh.handle.Debug("Refresh affect script execution, keys :", key, "error :", e)
}

func (dh *deleteHandle) DeleteSearch(table string, sfs []*gorm.StructField, indexKey []interface{}) (e error) {
	var keys []string
	for index := 0; index < len(sfs); index++ {
		key := dh.handle.JoinAffectKey(table, sfs[index].DBName)
		keys = append(keys, key)
		dh.handle.Debug("Add script delete affect cache Key :", key)
	}

	if len(indexKey) == 0 {
		_, e = dh.handle.redisClient.EvalSha(dh.delSha, keys).Result()
		dh.handle.Debug("Delete script execution, keys :", keys, "error :", e)
	} else {
		sks := make([]interface{}, 0, len(indexKey))
		for index := 0; index < len(indexKey); index++ {
			if fmt.Sprint(indexKey[index]) == "" {
				sks = append(sks, "")
			} else {
				sks = append(sks, fmt.Sprintf("_tag:%v_", indexKey[index]))
			}
		}
		_, e = dh.handle.redisClient.EvalSha(dh.delIndexSha, keys, sks...).Result()
		dh.handle.Debug("Delete script execution, keys :", keys, " argv:", sks, " error :", e)
	}
	return e
}

func (dh *deleteHandle) loadLua() {
	script := redis.NewScript(`
	for k,v in pairs(KEYS) do
		local delKeys = redis.call("HKEYS", v)
		for _, dv in pairs(delKeys) do
			redis.call("DEL", dv)
		end
		redis.call("DEL", v)
	end
	return true
`)
	sha, err := script.Load(dh.handle.redisClient).Result()
	if err != nil {
		panic(err)
	}
	dh.delSha = sha

	script = redis.NewScript(`
	for k,v in pairs(KEYS) do
		local delKeys = redis.call("HKEYS", v)
		for _, dv in pairs(delKeys) do
			for _, indexKey in pairs(ARGV) do
				if indexKey == "" and string.find(dv, "_tag:") == nil then
					redis.call("DEL", dv)
					redis.call("HDEL", v, dv)
				elseif indexKey ~= "" and string.find(dv, indexKey) ~= nil then
					redis.call("DEL", dv)
					redis.call("HDEL", v, dv)
				end
			end
		end
	end
	return true
`)
	sha, err = script.Load(dh.handle.redisClient).Result()
	if err != nil {
		panic(err)
	}
	dh.delIndexSha = sha

	script = redis.NewScript(`
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
	sha, err = script.Load(dh.handle.redisClient).Result()
	if err != nil {
		panic(err)
	}
	dh.refreshSearchSha = sha

	script = redis.NewScript(`
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
	sha, err = script.Load(dh.handle.redisClient).Result()
	if err != nil {
		panic(err)
	}
	dh.refreshAffectSha = sha
}
