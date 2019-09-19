package internal

import (
	"github.com/8treenet/gcache/option"
	"github.com/jinzhu/gorm"
)

const (
	_CACHE_UPDATE_BEFORE_PLUGIN = "CACHE:UPDATE_BEFORE_PLUGIN"
	_CACHE_UPDATE_AFTER_PLUGIN  = "CACHE:UPDATE_AFTER_PLUGIN"
)

type callUpdate struct {
	handle *Handle
}

func newCallUpdate(handle *Handle) *callUpdate {
	return &callUpdate{handle: handle}
}

func (cu *callUpdate) Bind() {
	cu.handle.db.Callback().Update().Before("gorm:update").Register(_CACHE_UPDATE_BEFORE_PLUGIN, cu.beforeInvoke)
	cu.handle.db.Callback().Update().After("gorm:update").Register(_CACHE_UPDATE_AFTER_PLUGIN, cu.afterInvoke)
}

func (cu *callUpdate) beforeInvoke(scope *gorm.Scope) {
	easyScope := newEasyScope(scope, cu.handle)
	if _, ok := easyScope.DB().Get(skipCache); ok || easyScope.opt.Level == option.LevelDisable {
		return
	}
	var primarys []interface{}
	var e error
	if !easyScope.PrimaryKeyZero() {
		//如果是单主键更新, 直接删除
		primarys = append(primarys, easyScope.PrimaryKeyValue())
	} else {
		easyScope = easyScope.QueryScope()
		if len(easyScope.condition.PrimaryValue) != 0 {
			primarys = easyScope.condition.PrimaryValue
		} else {
			//其他执行 select 主键 where
			primarys, e = easyScope.EasyPrimarys()
		}
	}

	if e != nil {
		return
	}

	scope.InstanceSet("easy_scope", easyScope)
	scope.InstanceSet("invalid_primarys", primarys)
}

func (cu *callUpdate) afterInvoke(scope *gorm.Scope) {
	//update 无影响 直接返回
	if scope.DB().RowsAffected <= 0 {
		return
	}

	var escope *easyScope
	var primarys []interface{}
	if inter, ok := scope.InstanceGet("easy_scope"); ok {
		escope = inter.(*easyScope)
	}
	if inter, ok := scope.InstanceGet("invalid_primarys"); ok {
		primarys = inter.([]interface{})
	}

	if escope == nil || len(primarys) == 0 {
		return
	}

	ds := true
	if escope.opt.Level == option.LevelModel {
		ds = false
	}

	writeRedis := func(delSearch bool) {
		cu.handle.NewDeleteHandle().delModle(escope.Table, primarys...)
		if delSearch {
			cu.handle.NewUpdateHandle().UpdateSearch(escope)
		}
	}

	if escope.opt.AsyncWrite {
		go writeRedis(ds)
	} else {
		writeRedis(ds)
	}
}
