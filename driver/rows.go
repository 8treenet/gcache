package driver

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

type fieldType byte
type fieldFlag uint16

type resultSet struct {
	columns     []cacheField
	columnNames []string
	done        bool
}

type cacheRows struct {
	mc     *driverConn
	rs     resultSet
	finish func()
}

type binaryRows struct {
	cacheRows
}

type textRows struct {
	cacheRows
	args []driver.NamedValue
}

func (rows *textRows) Columns() (s []string) {
	for index := 0; index < len(rows.args); index++ {
		s = append(s, rows.args[index].Name)
	}
	return
}

func (rows *cacheRows) ColumnTypeDatabaseTypeName(i int) string {
	return ""
}

func (rows *cacheRows) ColumnTypeNullable(i int) (nullable, ok bool) {
	return true, true
}

func (rows *cacheRows) ColumnTypePrecisionScale(i int) (int64, int64, bool) {
	return 0, 0, false
}

func (rows *cacheRows) ColumnTypeScanType(i int) reflect.Type {
	return rows.rs.columns[i].scanType()
}

func (rows *cacheRows) Close() (err error) {
	return nil
}

func (rows *cacheRows) HasNextResultSet() (b bool) {
	return true
}

func (rows *binaryRows) NextResultSet() error {
	return nil
}

func (rows *binaryRows) Next(dest []driver.Value) error {
	return nil
}

func (rows *textRows) NextResultSet() (err error) {
	return nil
}

func (rows *textRows) Next(dest []driver.Value) error {
	rows.readRow(dest)
	return nil
}

type cacheField struct {
	tableName string
	name      string
	length    uint32
	flags     fieldFlag
	fieldType fieldType
	decimals  byte
	charSet   uint8
}
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

func (mf *cacheField) scanType() reflect.Type {
	return reflect.TypeOf(NullTime{})
}

const (
	fieldTypeDecimal fieldType = iota
	fieldTypeTiny
	fieldTypeShort
	fieldTypeLong
	fieldTypeFloat
	fieldTypeDouble
	fieldTypeNULL
	fieldTypeTimestamp
	fieldTypeLongLong
	fieldTypeInt24
	fieldTypeDate
	fieldTypeTime
	fieldTypeDateTime
	fieldTypeYear
	fieldTypeNewDate
	fieldTypeVarChar
	fieldTypeBit
)

const (
	fieldTypeJSON fieldType = iota + 0xf5
	fieldTypeNewDecimal
	fieldTypeEnum
	fieldTypeSet
	fieldTypeTinyBLOB
	fieldTypeMediumBLOB
	fieldTypeLongBLOB
	fieldTypeBLOB
	fieldTypeVarString
	fieldTypeString
	fieldTypeGeometry
)

const (
	flagNotNULL fieldFlag = 1 << iota
	flagPriKey
	flagUniqueKey
	flagMultipleKey
	flagBLOB
	flagUnsigned
	flagZeroFill
	flagBinary
	flagEnum
	flagAutoIncrement
	flagTimestamp
	flagSet
	flagUnknown1
	flagUnknown2
	flagUnknown3
	flagUnknown4
)

func (rows *textRows) readRow(dest []driver.Value) error {
	data := []byte(fmt.Sprint(rows.args[0].Value))
	dest[0] = data
	return nil
}
