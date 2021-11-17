package internal

import (
	"errors"
	"reflect"
)

// InSlice .
func InSlice(array interface{}, item interface{}) bool {
	values := reflect.ValueOf(array)
	if values.Kind() != reflect.Slice {
		return false
	}

	size := values.Len()
	list := make([]interface{}, size)
	slice := values.Slice(0, size)
	for index := 0; index < size; index++ {
		list[index] = slice.Index(index).Interface()
	}

	for index := 0; index < len(list); index++ {
		if list[index] == item {
			return true
		}
	}
	return false
}

// SliceUp 切割1维数组，target:传入数组, result:返回二维数据,capacity:切割的容量
func SliceUp(target interface{}, result interface{}, capacity int) error {
	if capacity <= 0 {
		return nil
	}
	targetValue := reflect.ValueOf(target)
	resultValue := reflect.ValueOf(result)

	if targetValue.Kind() != reflect.Slice {
		return errors.New("target type error")
	}

	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return errors.New("result type error")
	}

	newValue := reflect.MakeSlice(resultValue.Elem().Type(), 0, 0)
	begin := 0
	for {
		j := begin + capacity
		if j > targetValue.Len() {
			j = targetValue.Len()
		}

		rangeSlice := targetValue.Slice(begin, j)
		newValue = reflect.Append(newValue, rangeSlice)
		begin = j
		if j == targetValue.Len() {
			break
		}
	}

	resultValue.Elem().Set(newValue)
	return nil
}
