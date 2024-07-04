package common

import (
	"fmt"
	"reflect"
)

func PrettyPrintStruct(data interface{}) {
	v := reflect.ValueOf(data)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
}
