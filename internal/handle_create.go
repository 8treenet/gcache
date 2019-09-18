package internal

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

func newCreateHandle(ch *Handle) *createHandle {
	return &createHandle{handle: ch}
}

type createHandle struct {
	handle *Handle
}

// CreateModel
func (ch *createHandle) CreateModel(table string, primary interface{}, model interface{}, expiration int) (e error) {
	var jsonModel JsonModel
	jsonModel.PK = fmt.Sprint(primary)
	jsonModel.Model = model
	buffer, e := json.Marshal(jsonModel)
	if e != nil {
		return
	}
	key := ch.handle.JoinModelKey(table, jsonModel.PK)
	e = ch.handle.redisClient.Set(key, string(buffer), time.Duration(expiration)*time.Second).Err()
	ch.handle.Debug("Set model cache Key :", key, "value :", string(buffer), "error :", e)
	return e
}

// CreateCountSearch
func (ch *createHandle) CreateCountSearch(table, key, field string, whereField []string, values []interface{}, expiration int, joins ...struct {
	ObjectField []string //使用的模型列
	Table       string   //表名
}) (e error) {
	field = ch.handle.JoinCountSecondKey(field)
	return ch.CreateSearch(table, key, field, whereField, values, expiration, joins...)
}

// CreateSearch
func (ch *createHandle) CreateSearch(table, key, field string, whereField []string, values []interface{}, expiration int, joins ...struct {
	ObjectField []string //使用的模型列
	Table       string   //表名
}) (e error) {
	return ch.createLuaSearch(table, key, field, whereField, values,expiration,joins...)
	//now := time.Now().Unix()
	//jsearch := &JsonSearch{UpdatedAt: now, Primarys: values}
	//buff, e := json.Marshal(jsearch)
	//if e != nil {
	//	return
	//}
	//
	//searchKey := ch.handle.JoinSearchKey(table, key)
	//e = ch.handle.redisClient.HSet(searchKey, field, buff).Err()
	//ch.handle.Debug("Set search cache Key :", searchKey, "field :", field, "value :", string(buff), "error :", e)
	//if e != nil {
	//	return
	//}
	//
	//e = ch.handle.redisClient.Expire(searchKey, time.Duration(expiration)*time.Second).Err()
	//if e != nil {
	//	return
	//}
	//
	//for index := 0; index < len(whereField); index++ {
	//	affectKey := ch.handle.JoinAffectKey(table, whereField[index])
	//	e = ch.handle.redisClient.HSet(affectKey, searchKey, now).Err()
	//	ch.handle.Debug("Set affect cache Key :", affectKey, "field :", searchKey, "value :", now, "error :", e)
	//	if e != nil {
	//		return e
	//	}
	//}
	//for index := 0; index < len(joins); index++ {
	//	for j := 0; j < len(joins[index].ObjectField); j++ {
	//		affectKey := ch.handle.JoinAffectKey(joins[index].Table, joins[index].ObjectField[j])
	//		e = ch.handle.redisClient.HSet(affectKey, searchKey, now).Err()
	//		ch.handle.Debug("Set affect cache Key :", affectKey, "field :", searchKey, "value :", now, "error :", e)
	//		if e != nil {
	//			return e
	//		}
	//	}
	//}
	//return
}

// CreateSearch
func (ch *createHandle) createLuaSearch(table, key, field string, whereField []string, values []interface{}, expiration int, joins ...struct {
	ObjectField []string //使用的模型列
	Table       string   //表名
}) (e error) {
	now := time.Now().Unix()
	var keys []string
	var argv []interface{}
	jsearch := &JsonSearch{UpdatedAt: now, Primarys: values}
	buff, e := json.Marshal(jsearch)
	if e != nil {
		return
	}

	searchKey := ch.handle.JoinSearchKey(table, key)
	keys = append(keys, searchKey)
	argv = append(argv, field, string(buff), expiration, now)
	ch.handle.Debug("Add script set search cache key :", searchKey, "field :", field, "value :", string(buff), "error :", nil)

	for index := 0; index < len(whereField); index++ {
		affectKey := ch.handle.JoinAffectKey(table, whereField[index])
		keys = append(keys, affectKey)
		ch.handle.Debug("Add script set affect cache key :", affectKey, "field :", searchKey, "value :", now, "error :", nil)
	}
	for index := 0; index < len(joins); index++ {
		for j := 0; j < len(joins[index].ObjectField); j++ {
			affectKey := ch.handle.JoinAffectKey(joins[index].Table, joins[index].ObjectField[j])
			keys = append(keys, affectKey)
			ch.handle.Debug("Add script set affect cache key :", affectKey, "field :", searchKey, "value :", now, "error :", e)
		}
	}
	script := redis.NewScript(`
	for k,v in pairs(KEYS) do
		if k > 1 then
			local x = k-1
			redis.call("HSET", v, KEYS[1], ARGV[4])
		else 
			redis.call("HSET", v, ARGV[1], ARGV[2])
			redis.call("EXPIRE", v, ARGV[3])
		end

	end
	return true
`)
	_, e = script.Run(ch.handle.redisClient, keys, argv...).Result()
	ch.handle.Debug("Create script execution, keys :", keys, "value:", argv, "error :", e)
	return e
}
