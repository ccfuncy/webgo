package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

var Validator = defaultValidator{}

type StructValidator interface {
	//结构体验证器，如果错误返回对应的信息
	ValidateStruct(any) error
	//返回对应使用的验证器
	Engine() any
}

type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	of := reflect.ValueOf(obj)
	switch of.Kind() {
	case reflect.Pointer:
		return d.ValidateStruct(of.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(of)
	case reflect.Slice, reflect.Array:
		count := of.Len()
		sliceValidationError := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(of.Index(i).Interface()); err != nil {
				sliceValidationError = append(sliceValidationError, err)
			}
		}
		if len(sliceValidationError) == 0 {
			return nil
		}
		return sliceValidationError
	}
	return nil
}

func (d *defaultValidator) Engine() any {
	d.LazyInit()
	return d.validate
}

func (d *defaultValidator) validateStruct(obj any) error {
	d.LazyInit()
	return validator.New().Struct(obj)
}

func (d *defaultValidator) LazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]:[%s]", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				b.WriteString("\n")
				fmt.Fprintf(&b, "[%d]:[%s]", i, err[i].Error())
			}
		}
		return b.String()
	}
}

func validate(obj any) error {
	return Validator.ValidateStruct(obj)
}

func validateParam(obj any, decoder *json.Decoder) error {
	//解析成map，根据map中的key比对
	// valueof专注于对对象实例的读写
	valueOf := reflect.ValueOf(obj)
	//判断是否为指针类型
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("no")
	}
	//获取目标元素类型
	elem := valueOf.Elem().Interface()
	//拿到真实类型的valueof
	of := reflect.ValueOf(elem)
	//判断
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(of, obj, decoder)
	case reflect.Slice, reflect.Array:
		elem := of.Type().Elem()
		if elem.Kind() == reflect.Struct {
			return checkParamSlice(elem, obj, decoder)
		}
	default:
		_ = decoder.Decode(obj)
	}
	return nil
}

func checkParamSlice(of reflect.Type, obj any, decoder *json.Decoder) error {
	mapValue := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		field := of.Field(i)
		name := field.Name
		// json tag
		jsonName := field.Tag.Get("json")
		// require tag
		require := field.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		for _, v := range mapValue {
			_, ok := v[name]
			if !ok && require == "require" {
				return errors.New(fmt.Sprintf("filed [%s] is not exist!", name))
			}
		}
	}
	//赋值
	data, err := json.Marshal(mapValue)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	return nil
}

func checkParam(of reflect.Value, obj any, decoder *json.Decoder) error {
	mapValue := make(map[string]interface{})
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		name := field.Name
		// json tag
		jsonName := field.Tag.Get("json")
		// require tag
		require := field.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		_, ok := mapValue[name]
		if !ok && require == "require" {
			return errors.New(fmt.Sprintf("filed [%s] is not exist!", name))
		}
	}
	//赋值
	data, err := json.Marshal(mapValue)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	return nil
}
