package common

import (
	"fmt"
	"reflect"
)

func PrettyPrintStruct(data interface{}) {
	v := reflect.ValueOf(data)

	// If it's a pointer, get the underlying element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if it's a struct
	if v.Kind() != reflect.Struct {
		fmt.Println("Not a struct or pointer to struct")
		return
	}

	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
}
