package internal

import "github.com/jinzhu/gorm"

type search struct {
	db               *gorm.DB
	whereConditions  []map[string]interface{}
	orConditions     []map[string]interface{}
	notConditions    []map[string]interface{}
	havingConditions []map[string]interface{}
	joinConditions   []map[string]interface{}
	initAttrs        []interface{}
	assignAttrs      []interface{}
	selects          map[string]interface{}
	omits            []string
	orders           []interface{}
	preload          []searchPreload
	offset           interface{}
	limit            interface{}
	group            string
	tableName        string
	raw              bool
	Unscoped         bool
	ignoreOrderQuery bool
}

type searchPreload struct {
	schema     string
	conditions []interface{}
}

func (s *search) clone() *search {
	clone := *s
	return &clone
}
