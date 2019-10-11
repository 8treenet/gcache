package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/go-redis/redis"
)

func newQueryHandle(ch *Handle) *queryHandle {
	return &queryHandle{handle: ch}
}

type queryHandle struct {
	handle *Handle
}

// ByPrimary 通过主键查询 模型列表
func (q *queryHandle) ByPrimary(scope *easyScope, primarys ...interface{}) (objs []interface{}, e error) {
	models, err := q.byCache(scope, scope.valueType, primarys...)
	if err != nil {
		e = err
		return
	}
	for index := 0; index < len(models); index++ {
		if models[index].Model != nil {
			objs = append(objs, models[index].Model)
		}
	}

	if len(models) == len(primarys) {
		return
	}

	missPrimarys := []interface{}{}
	for index := 0; index < len(primarys); index++ {
		miss := false
		for j := 0; j < len(models); j++ {
			if models[j].PK == fmt.Sprint(primarys[index]) {
				miss = true
				break
			}
		}

		if !miss {
			missPrimarys = append(missPrimarys, primarys[index])
		}
	}

	value := reflect.MakeSlice(reflect.SliceOf(scope.valueType), 0, 0)
	value = reflect.New(value.Type())
	newScope := scope.DB().NewScope(value.Interface())
	s := (*search)(unsafe.Pointer(newScope.Search))
	s.whereConditions = nil
	newScope.Search.Where(scope.PrimaryKey()+" in(?)", missPrimarys)
	newScope.Search.Table(scope.Table)
	query := newScope.DB().Callback().Query().Get("gorm:query")
	query(newScope)
	preload := newScope.DB().Callback().Query().Get("gorm:preload")
	preload(newScope)
	after_query := newScope.DB().Callback().Query().Get("gorm:after_query")
	after_query(newScope)
	e = newScope.DB().Error
	if e != nil {
		return
	}

	rows := value.Elem()
	var create *createHandle
	create = newCreateHandle(q.handle)
	if rows.Len() == 0 && scope.opt.PenetrationSafe {
		//防穿透，填入空数据
		for index := 0; index < len(missPrimarys); index++ {
			create.CreateModel(scope.Table, missPrimarys[index], nil, scope.opt.Expires)
		}
	}

	pkFieldName := scope.PrimaryFieldName()
	for index := 0; index < rows.Len(); index++ {
		row := rows.Index(index)
		pk := row.FieldByName(pkFieldName).Interface()
		if pk == nil {
			continue
		}
		create.CreateModel(scope.Table, pk, row.Interface(), scope.opt.Expires)
		objs = append(objs, row.Interface())
	}
	return
}

// BySearch 通过条件查询主键列表
func (q *queryHandle) BySearch(scope *easyScope) (primarys []interface{}, e error) {
	jsearch, e := q.getSearchPrimarys(scope.Table, scope.condition.SQLKey, scope.condition.SQLValue, scope.indexKeys)
	if e != nil {
		return
	}
	if jsearch != nil && jsearch.Timeout > time.Now().Unix() {
		primarys = jsearch.Primarys
		return
	}

	rows, err := scope.EasyPrimarys()
	if err != nil {
		e = err
		return
	}

	if len(rows) > 0 || scope.opt.PenetrationSafe {
		create := newCreateHandle(q.handle)
		if e = create.CreateSearch(scope.Table, scope.condition.SQLKey, scope.condition.SQLValue, scope.condition.ObjectField, rows, scope.opt.Expires, scope.indexKeys, scope.joinsCondition...); e != nil {
			return
		}
	}
	primarys = rows
	return
}

//ByCount 通过条件查询count
func (q *queryHandle) ByCount(scope *easyScope) (count int, e error) {
	jsearch, e := q.getSearchCount(scope.Table, scope.condition.SQLKey, scope.condition.SQLCountValue, scope.indexKeys)
	if e != nil {
		return
	}

	if jsearch != nil && jsearch.Timeout > time.Now().Unix() && len(jsearch.Primarys) > 0 {
		count, _ = strconv.Atoi(fmt.Sprint(jsearch.Primarys[0]))
		return
	}

	count, e = scope.EasyCount()
	if e != nil {
		return
	}

	if count > 0 || scope.opt.PenetrationSafe {
		create := newCreateHandle(q.handle)
		if e = create.CreateCountSearch(scope.Table, scope.condition.SQLKey, scope.condition.SQLCountValue, scope.condition.ObjectField, []interface{}{count}, scope.opt.Expires, scope.indexKeys, scope.joinsCondition...); e != nil {
			return
		}
	}
	return
}

// getSearchPrimarys
func (q *queryHandle) getSearchPrimarys(table string, key string, field string, indexKeys []interface{}) (jsearch *JsonSearch, e error) {
	searchKey := q.handle.JoinSearchKey(table, key, indexKeys)
	value, e := q.handle.redisClient.HGet(searchKey, field).Result()
	if e == redis.Nil {
		e = nil
	}
	q.handle.Debug("Get search cache Key :", searchKey, "field :", field, "hit:", value != "", "error :", e)

	if e != nil || value == "" {
		return
	}

	jsearch = new(JsonSearch)
	if e = json.Unmarshal([]byte(value), jsearch); e != nil {
		return
	}
	return
}

// getSearchPrimarys
func (q *queryHandle) getSearchCount(table string, key string, field string, indexKeys []interface{}) (jsearch *JsonSearch, e error) {
	field = q.handle.JoinCountSecondKey(field)
	serachKey := q.handle.JoinSearchKey(table, key, indexKeys)
	value, e := q.handle.redisClient.HGet(serachKey, field).Result()
	if e == redis.Nil {
		e = nil
	}
	q.handle.Debug("Get search (count*) cache Key :", serachKey, "field :", field, "hit:", value != "", "error :", e)

	if e != nil || value == "" {
		return
	}

	jsearch = new(JsonSearch)
	if e = json.Unmarshal([]byte(value), jsearch); e != nil {
		return
	}
	return
}

func (q *queryHandle) byCache(scope *easyScope, t reflect.Type, primarys ...interface{}) (result []*JsonModel, e error) {
	keys := q.handle.JoinModelKeys(scope.Table, primarys...)
	list, e := q.handle.redisClient.MGet(keys...).Result()
	if e != nil {
		q.handle.Debug("Get model cache Key:", keys, "hit:[]", "error:", e)
		return
	}

	hits := []bool{}
	for index := 0; index < len(list); index++ {
		str, ok := list[index].(string)
		if !ok {
			hits = append(hits, false)
			continue
		}

		jsonmodel := new(JsonModel)
		jsonmodel.Model = reflect.New(t).Interface()
		json.Unmarshal([]byte(str), jsonmodel)
		result = append(result, jsonmodel)
		hits = append(hits, true)
	}
	q.handle.Debug("Get model cache Key:", keys, "hit:", hits, "error:", e)
	return
}
