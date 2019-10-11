package internal

import "github.com/jinzhu/gorm"

func newUpdateHandle(handle *Handle) *updateHandle {
	return &updateHandle{handle: handle}
}

type updateHandle struct {
	handle *Handle
}

// UpdateSearch 更新查询缓存
func (uh *updateHandle) UpdateSearch(scope *easyScope) {
	filds := scope.GetStructFields()
	var invalidFilds []*gorm.StructField
	updateAttrs, ok := scope.sourceScope.InstanceGet("gorm:update_attrs")
	if !ok {
		return
	}

	updateMap := updateAttrs.(map[string]interface{})
	for k, _ := range updateMap {
		if k == "updated_at" {
			delete(updateMap, "updated_at")
			continue
		}
	}

	for index := 0; index < len(filds); index++ {
		for k, _ := range updateMap {
			if k == filds[index].DBName {
				invalidFilds = append(invalidFilds, filds[index])
			}
		}
	}

	//只删除影响字段的查询缓存
	newDeleteHandle(uh.handle).DeleteSearch(scope.Table, invalidFilds, scope.indexKeys)
}
