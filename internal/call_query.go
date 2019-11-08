package internal

import (
	"fmt"
	"reflect"

	"github.com/8treenet/gcache/option"

	"github.com/jinzhu/gorm"
)

const (
	_CACHE_QUERY_PLUGIN = "CACHE:QUERY_PLUGIN"
)

func newCallQuery(handle *Handle) *callQuery {
	return &callQuery{handle: handle}
}

type callQuery struct {
	handle      *Handle
	singleGroup Group
}

// Bind
func (c *callQuery) Bind() {
	c.handle.db.Callback().Query().Before("gorm:query").Register(_CACHE_QUERY_PLUGIN, c.invoke)
}

type singleQuery struct {
	Skip   bool
	Models []interface{}
}

// invoke
func (c *callQuery) invoke(scope *gorm.Scope) {
	easyScope := newEasyScope(scope, c.handle)
	if c.pass(easyScope) {
		return
	}
	easyScope = easyScope.QueryScope()
	if len(easyScope.condition.PrimaryValue) == 0 && len(easyScope.condition.ObjectField) == 0 && len(easyScope.joinsModels) == 0 {
		return
	}

	v, _, _ := c.singleGroup.Do(easyScope.condition.SQLKey+easyScope.condition.SQLValue+fmt.Sprint(easyScope.condition.PrimaryValue), func() (i interface{}, e error) {
		var s singleQuery
		if ok, list := c.byPrimary(easyScope); ok {
			s.Models = list
			s.Skip = ok
			return s, nil
		}
		if ok, list := c.bySearch(easyScope); ok {
			s.Models = list
			s.Skip = ok
			return s, nil
		}
		return s, nil
	})

	s, ok := v.(singleQuery)
	if ok && s.Skip {
		c.setIndirectValue(easyScope, s.Models)
		scope.SkipLeft()
		return
	}
}

func (c *callQuery) pass(es *easyScope) bool {
	if _, ok := es.DB().Get(skipCache); ok || es.isJoinSkip() || es.isSelectSkip() || es.forgeSearch.group != "" || es.forgeSearch.havingConditions != nil {
		return true
	}
	if _, ok := es.InstanceGet("gorm:started_transaction"); ok {
		return true
	}

	if es.opt.Level == option.LevelDisable {
		return true
	}

	return false
}

func (c *callQuery) byPrimary(es *easyScope) (ok bool, list []interface{}) {
	ok = false
	if len(es.condition.PrimaryValue) == 0 || len(es.condition.ObjectField) > 0 {

		return
	}
	value := es.IndirectValue()
	if !value.CanSet() {
		return
	}
	models, err := c.handle.NewQueryHandle().ByPrimary(es, es.condition.PrimaryValue...)
	if err != nil {
		return
	}

	list = models
	ok = true
	return
}

func (c *callQuery) bySearch(es *easyScope) (ok bool, list []interface{}) {
	ok = false

	if es.opt.Level < option.LevelSearch || len(es.condition.PrimaryValue) > 0 || (len(es.condition.ObjectField) <= 0 && len(es.joinsModels) == 0) {
		return
	}
	revalue := es.IndirectValue()
	if !revalue.CanSet() {
		return
	}
	qh := c.handle.NewQueryHandle()
	pks, err := qh.BySearch(es)
	if err != nil {
		return
	}
	if len(pks) == 0 {
		ok = true
		return
	}

	models, err := qh.ByPrimary(es, pks...)
	if err != nil {
		return
	}

	if len(models) > 1 {
		models = c.sortModels(es, pks, models)
	}

	ok = true
	list = models
	return
}

func (c *callQuery) sortModels(es *easyScope, primarys []interface{}, inModels []interface{}) (outModels []interface{}) {
	m := make(map[string]interface{})
	pfname := es.PrimaryFieldName()

	for index := 0; index < len(inModels); index++ {
		value := reflect.ValueOf(inModels[index])
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		fieldValue := value.FieldByName(pfname)
		if fieldValue.IsValid() {
			m[fmt.Sprint(fieldValue.Interface())] = inModels[index]
		}
	}

	for index := 0; index < len(primarys); index++ {
		model, ok := m[fmt.Sprint(primarys[index])]
		if ok {
			outModels = append(outModels, model)
		}
	}
	return
}

func (c *callQuery) setIndirectValue(es *easyScope, models []interface{}) {
	value := es.IndirectValue()
	if value.Kind() == reflect.Slice {
		value.Set(reflect.MakeSlice(value.Type(), 0, len(models)))
		for index := 0; index < len(models); index++ {
			model := reflect.ValueOf(models[index])
			if model.Kind() == reflect.Ptr {
				model = model.Elem()
			}
			value = reflect.Append(value, model)
		}
		es.IndirectValue().Set(value)
		return
	} else if len(models) > 0 {
		model := reflect.ValueOf(models[0])
		if model.Kind() == reflect.Ptr {
			model = model.Elem()
		}
		es.IndirectValue().Set(model)
		return
	}
	es.sourceScope.Err(gorm.ErrRecordNotFound)
}
