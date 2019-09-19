package internal

import (
	"github.com/8treenet/gcache/option"
	"github.com/jinzhu/gorm"
)

type plugin struct {
	logMode    bool
	db         *gorm.DB
	defaultOpt *option.DefaultOption
	handle     *Handle
}

func InjectGorm(db *gorm.DB, opt *option.DefaultOption,redisOption *option.RedisOption) *plugin {
	cp := new(plugin)
	opt.Init()
	cp.db = db
	cp.defaultOpt = opt

	handle := newHandleManager(db, cp, redisOption)
	handle.registerCall()
	go handle.RefreshRun()
	cp.handle = handle
	return cp
}

// FlushDB
func (cp *plugin) FlushDB() error {
	return cp.handle.NewDeleteHandle().flushDB()
}

// DeleteModel
func (cp *plugin) DeleteModel(model interface{}, primarys ...interface{}) error {
	table := cp.db.NewScope(model).TableName()
	return cp.handle.NewDeleteHandle().delModle(table, primarys...)
}

// DeleteSearch
func (cp *plugin) DeleteSearch(model interface{}) error {
	scope := cp.db.NewScope(model)
	return cp.handle.NewDeleteHandle().DeleteSearchByScope(newEasyScope(scope, cp.handle))
}

// SkipCache
func (cp *plugin) SkipCache() *gorm.DB {
	return cp.db.New().InstantSet(skipCache, true)
}

// UseModels
func (cp *plugin) UseModels(models ...interface{}) *gorm.DB {
	if len(models) == 0 {
		panic("models empty")
	}

	return cp.db.New().InstantSet(whereModelsSearch, models)
}

// Debug
func (cp *plugin) Debug() {
	cp.handle.debug = true
}
