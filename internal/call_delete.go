package internal

import (
	"github.com/8treenet/gcache/option"
	"github.com/jinzhu/gorm"
)

const (
	_CACHE_DELETE_BEFORE_PLUGIN = "CACHE:DELETE_BEFORE_PLUGIN"
	_CACHE_DELETE_AFTER_PLUGIN  = "CACHE:DELETE_AFTER_PLUGIN"
)

//gorm:delete
func newCallDelete(handle *Handle) *callDelete {
	return &callDelete{handle: handle}
}

type callDelete struct {
	handle *Handle
}

func (del *callDelete) Bind() {
	del.handle.db.Callback().Delete().Before("gorm:delete").Register(_CACHE_DELETE_BEFORE_PLUGIN, del.beforeInvoke)
	del.handle.db.Callback().Delete().After("gorm:delete").Register(_CACHE_DELETE_AFTER_PLUGIN, del.afterInvoke)
}

// beforeInvoke
func (del *callDelete) beforeInvoke(scope *gorm.Scope) {
	easyScope := newEasyScope(scope, del.handle)
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

// afterInvoke
func (del *callDelete) afterInvoke(scope *gorm.Scope) {
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
		delhandle := del.handle.NewDeleteHandle()
		delhandle.delModle(escope.Table, primarys...)
		if delSearch {
			delhandle.DeleteSearchByScope(escope)
		}
	}

	if escope.opt.AsyncWrite {
		go writeRedis(ds)
	} else {
		writeRedis(ds)
	}
}
