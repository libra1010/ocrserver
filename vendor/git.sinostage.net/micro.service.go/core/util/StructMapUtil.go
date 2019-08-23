package koloCore

import (
	"fmt"
	"github.com/goinggo/mapstructure"
	"reflect"
	"strconv"
)

type structMapUtil uint

const StructMapUtil structMapUtil = 1

func (this structMapUtil) StructToMap(obj interface{}) map[string]string {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)

	var data = make(map[string]string)
	for i := 0; i < obj1.NumField(); i++ {

		data[obj1.Field(i).Name] = this.AnyToString(obj2.Field(i).Interface())
	}
	return data
}

func (this structMapUtil) AnyToString(obj interface{}) string {
	vt := reflect.TypeOf(obj)

	switch vt.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		{
			var value int64
			value = obj.(int64)
			return strconv.FormatInt(value, 10)
		}
	case reflect.Float64, reflect.Float32:
		{
			var value float64
			value = obj.(float64)
			return strconv.FormatFloat(value, 'f', -1, 64)
		}

	case reflect.Map, reflect.Array, reflect.Slice:
		{
			return ""
		}

	default:
		return obj.(string)
	}
}

func (this structMapUtil) MapToStruct(mapArgs map[string]interface{}, result interface{}) error {

	err := mapstructure.Decode(mapArgs, result)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (this structMapUtil) MapStringToStruct(mapArgs map[string]string, result interface{}) error {
	err := mapstructure.Decode(mapArgs, result)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
