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
	return &deleteHandle{handle: handle}
}

type deleteHandle struct {
	handle *Handle
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

func (dh *deleteHandle) delSearchByFields(table string, sfs []*gorm.StructField) error {
	now := time.Now()
	for index := 0; index < len(sfs); index++ {
		var delTimeout []string
		var delSearch []string
		key := dh.handle.JoinAffectKey(table, sfs[index].DBName)
		m, err := dh.handle.redisClient.HGetAll(key).Result()
		if err != nil {
			continue
		}

		for searchKey, v := range m {
			fieldUnix, fieldUnixE := strconv.Atoi(v)
			if fieldUnixE == nil && now.Sub(time.Unix(int64(fieldUnix), 0)).Hours() > 48 {
				delTimeout = append(delTimeout, searchKey)
			}
			delSearch = append(delSearch, searchKey)
		}

		if len(delSearch) > 0 {
			err = dh.handle.redisClient.Del(delSearch...).Err()
			dh.handle.Debug("Delete search cache Key :", delSearch, "error :", err)
			if err != nil {
				return err
			}
		}

		if len(delTimeout) > 0 {
			dh.handle.Debug("Delete affect cache Key :", delTimeout, "error :", err)
			err = dh.handle.redisClient.HDel(key, delTimeout...).Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dh *deleteHandle) delSearchByScope(scope *easyScope) error {
	table := scope.Table
	sfs := scope.GetStructFields()
	return dh.delSearchByFields(table, sfs)
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
