package internal

import (
	"github.com/8treenet/gcache/option"
	"github.com/jinzhu/gorm"
)

const (
	_CACHE_CREATE_BEFORE_PLUGIN = "CACHE:CREATE_BEFORE_PLUGIN"
	_CACHE_CREATE_AFTER_PLUGIN  = "CACHE:CREATE_AFTER_PLUGIN"
)

func newCallCreate(handle *Handle) *callCreate {
	return &callCreate{handle: handle}
}

type callCreate struct {
	handle *Handle
}

// Bind
func (create *callCreate) Bind() {
	create.handle.db.Callback().Create().Before("gorm:create").Register(_CACHE_CREATE_BEFORE_PLUGIN, create.beforeInvoke)
	create.handle.db.Callback().Create().After("gorm:create").Register(_CACHE_CREATE_AFTER_PLUGIN, create.afterInvoke)
}

// beforeInvoke
func (create *callCreate) beforeInvoke(scope *gorm.Scope) {
	easyScope := newEasyScope(scope, create.handle)
	if _, ok := easyScope.DB().Get(skipCache); ok || easyScope.opt.Level == option.LevelDisable {
		return
	}
	scope.InstanceSet("easy_scope", easyScope)
}

// afterInvoke
func (create *callCreate) afterInvoke(scope *gorm.Scope) {
	//update 无影响 直接返回
	if scope.DB().RowsAffected <= 0 {
		return
	}
	var escope *easyScope
	if inter, ok := scope.InstanceGet("easy_scope"); ok {
		escope = inter.(*easyScope)
	}

	if escope == nil {
		return
	}

	ds := true
	//只开启模型缓存
	if escope.opt.Level == option.LevelModel {
		ds = false
	}

	writeRedis := func(delSearch bool) {
		delHandle := create.handle.NewDeleteHandle()
		if delSearch {
			delHandle.DeleteSearchByScope(escope)
		}

		if escope.opt.PenetrationSafe {
			delHandle.delModle(escope.Table, escope.PrimaryKeyValue())
		}
	}

	if escope.opt.AsyncWrite {
		go writeRedis(ds)
	} else {
		writeRedis(ds)
	}
}
