package internal

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
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
	handle *Handle
	delLuaSha string
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
	return dh.DeleteSearch(table, sfs)
}

func (dh *deleteHandle) refresh(t reflect.Type) {
	value := reflect.New(t)
	newScope := dh.handle.db.NewScope(value.Interface())
	table := newScope.TableName()
	sfs := newScope.GetStructFields()
	now := time.Now()

	for index := 0; index < len(sfs); index++ {
		var delTimeout []string
		var checkSearch []string
		key := dh.handle.JoinAffectKey(table, sfs[index].DBName)
		m, err := dh.handle.redisClient.HGetAll(key).Result()
		if err != nil {
			continue
		}

		for searchKey, v := range m {
			fieldUnix, fieldUnixE := strconv.Atoi(v)
			if fieldUnixE == nil && now.Sub(time.Unix(int64(fieldUnix), 0)).Hours() > 48.0 {
				delTimeout = append(delTimeout, searchKey)
			}
			checkSearch = append(checkSearch, searchKey)
		}

		if len(checkSearch) > 0 {
			for index := 0; index < len(checkSearch); index++ {
				dh.timeoutSearch(checkSearch[index], now)
			}
		}

		if len(delTimeout) > 0 {
			err = dh.handle.redisClient.HDel(key, delTimeout...).Err()
			dh.handle.Debug("Delete affect cache Key :", delTimeout, "error :", err)
			if err != nil {
				return
			}
		}
	}
}

func (dh *deleteHandle) timeoutSearch(searchKey string, now time.Time) {
	cursor := uint64(0)
	var e error
	var keys []string
	var batchKeys []string

	for index := 0; index < 20; index++ {
		var list []string
		list, cursor, e = dh.handle.redisClient.HScan(searchKey, cursor, "*", 500).Result()
		keys = append(keys, list...)
		if e != nil || cursor == 0 {
			break
		}
	}

	process := func(fields []string) {
		var delfields []string
		for index := 0; index < len(fields); index++ {
			time.Sleep(10 * time.Millisecond)
			value, e := dh.handle.redisClient.HGet(searchKey, fields[index]).Result()
			if e == redis.Nil {
				e = nil
			}
			if e != nil || value == "" {
				continue
			}

			jsearch := new(JsonSearch)
			if e = json.Unmarshal([]byte(value), jsearch); e != nil {
				return
			}

			if now.Sub(time.Unix(int64(jsearch.UpdatedAt), 0)).Minutes() > 30.0 {
				delfields = append(delfields, fields[index])
			}
		}

		if len(delfields) <= 0 {
			return
		}

		err := dh.handle.redisClient.HDel(searchKey, delfields...).Err()
		dh.handle.Debug("Delete search cache Key :", searchKey, "field :", delfields, "error :", err)
	}

	for index := 0; index < len(keys); index++ {
		batchKeys = append(batchKeys, keys[index])
		if len(batchKeys) > 100 {
			process(batchKeys)
			batchKeys = []string{}
		}
	}

	if len(batchKeys) > 0 {
		process(batchKeys)
	}
}


func (dh *deleteHandle) DeleteSearch(table string, sfs []*gorm.StructField) error {
	var keys []string
	for index := 0; index < len(sfs); index++ {
		key := dh.handle.JoinAffectKey(table, sfs[index].DBName)
		keys = append(keys, key)
		dh.handle.Debug("Add script delete affect cache Key :", key)
	}

	_, e := dh.handle.redisClient.EvalSha(dh.delLuaSha, keys).Result()
	dh.handle.Debug("Delete script execution, keys :", keys, "error :", e)
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
	dh.delLuaSha = sha
}