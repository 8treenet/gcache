package internal

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/8treenet/gcache/driver"
	"github.com/8treenet/gcache/option"
	"github.com/jinzhu/gorm"
)

const (
	CACHE_ROW_PLUGIN = "CACHE:ROW_PLUGIN"
)

func newCallRow(handle *Handle) *callRow {
	db, _ := sql.Open("cache_plugin", "")
	return &callRow{handle: handle, driver: db}
}

type callRow struct {
	driver      *sql.DB
	handle      *Handle
	singleGroup Group
}

func (c *callRow) Bind() {
	c.handle.db.Callback().RowQuery().Before("gorm:row_query").Register(CACHE_ROW_PLUGIN, c.invoke)
}

func (c *callRow) invoke(scope *gorm.Scope) {
	easyScope := newEasyScope(scope, c.handle)
	if c.pass(easyScope) {
		return
	}
	easyScope = easyScope.QueryScope()

	var result interface{}
	var ok bool
	var rowResult *gorm.RowQueryResult
	if result, ok = scope.InstanceGet("row_query_result"); !ok {
		return
	}
	if rowResult, ok = result.(*gorm.RowQueryResult); !ok {
		return
	}

	count, e, _ := c.singleGroup.Do(easyScope.condition.SQLKey+easyScope.condition.SQLCountValue+fmt.Sprint(easyScope.condition.PrimaryValue), func() (interface{}, error) {
		return c.QueryCount(easyScope)
	})

	num, ok := count.(int)
	if e == nil && ok {
		rowResult.Row = c.driver.QueryRow("", num)
		scope.SkipLeft()
	}
}

func (c *callRow) pass(es *easyScope) bool {
	if _, ok := es.DB().Get(skipCache); ok || es.isJoinSkip() || es.forgeSearch.group != "" || es.forgeSearch.havingConditions != nil {
		return true
	}
	if _, ok := es.sourceScope.InstanceGet("gorm:started_transaction"); ok {
		return true
	}
	if _, ok := es.Get("cache:easy_count"); ok {
		return true
	}

	if es.opt.Level == option.LevelDisable || es.opt.Level == option.LevelModel {
		return true
	}

	if es.forgeSearch.selects == nil || es.forgeSearch.selects["query"] != "count(*)" {
		return true
	}
	return false
}

// QueryCount
func (c *callRow) QueryCount(es *easyScope) (int, error) {
	if len(es.condition.PrimaryValue) > 0 || (len(es.condition.ObjectField) <= 0 && len(es.joinsModels) == 0) {
		return 0, errors.New("len(es.condition.Primary) > 0 || len(es.condition.Field) <= 0")
	}

	qh := c.handle.NewQueryHandle()
	count, err := qh.ByCount(es)
	return count, err
}
