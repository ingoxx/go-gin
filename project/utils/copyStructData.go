package utils

import (
	"errors"
	"reflect"
)

func CopyStruct(src, dst interface{}) (err error) {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	if srcType.Kind() != reflect.Ptr || dstType.Kind() != reflect.Ptr {
		return errors.New("not the expected pointer type")
	}

	if srcType.Elem().Kind() != reflect.Struct && dstType.Elem().Kind() != reflect.Struct {
		return errors.New("not the expected struct type")
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	srcFieldList := srcType.Elem().NumField()
	dstFieldList := dstType.Elem().NumField()

	srcElem := srcVal.Elem()
	dstElem := dstVal.Elem()

	for i := 0; i < srcFieldList; i++ {
		for t := 0; t < dstFieldList; t++ {
			if srcType.Elem().Field(i).Name == dstType.Elem().Field(t).Name && srcVal.Elem().Field(i).Kind() == dstVal.Elem().Field(t).Kind() {
				switch dstVal.Elem().Field(t).Kind() {
				case reflect.String:
					dstElem.Field(t).SetString(srcElem.Field(i).String())
				case reflect.Int:
					dstElem.Field(t).SetInt(srcElem.Field(i).Int())
				case reflect.Int32:
					dstElem.Field(t).SetInt(srcElem.Field(i).Int())
				default:
					panic("unhandled default case")
				}
			}
		}
	}

	return
}
