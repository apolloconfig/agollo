package objectutils

import "reflect"

func IsNil(object interface{}) bool {

	return isNil(object)
}

func IsNotNil(object interface{}) bool {
	return !isNil(object)
}


func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}

