package gtpl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gtpl/types"
)

func parseGlobalStructVal(tagName string, values *reflect.Value) string {
	if arr := strings.Split(tagName, "."); len(arr) > 1 {
		if values.Kind().String() != "slice" {
			// fmt.Println(values.Elem().Type().Name())
			// if structName := values.Elem().Type().Name(); structName == case2CamelS(arr[0]) {
			rv := values.Elem().FieldByName(case2CamelS(arr[1]))
			vk := rv.Kind().String()
			if vk == "int64" {
				return fmt.Sprintf("%d", rv.Int())
			}
			if vk == "string" {
				return rv.String()
			}
			// fmt.Println(tagName, vk, rv.Elem().Type())
			return parseGlobalStructVal(strings.Join(arr[1:], "."), &rv)
			// }
		}
	}

	return values.String()
}

// parseVal 解析变量
func parseVal(tagName string, vars []*reflect.Value, thisv *reflect.Value) string {
	vlen := len(vars)
	if thisv == nil && vlen > 0 {
		thisv = vars[vlen-1]
	}

	thisvKind := thisv.Kind().String()
	if thisvKind == "map" { // 处理 map 格式的数据 这类数据通常由 sql 标签得到
		if tagValue := thisv.MapIndex(reflect.ValueOf(tagName)); tagValue.IsValid() {
			return tagValue.String()
		}
		return parseVal(tagName, vars, thisv)
	}

	// 取当前作用域上层数据
	// 应找到上层数据非数组形式的反射结构体或 map
	// 仅允许取两级数据，否则将对性能造成一定影响
	// if arr := strings.Split(tagName, "."); len(arr) > 1 {
	// 	for i := vlen - 1; i >= 0; i-- {
	// 		if vs := vars[i]; vs.Kind().String() != "slice" {
	// 			if structName := vs.Elem().Type().Name(); structName == case2CamelS(arr[0]) {
	// 				rv := vs.Elem().FieldByName(case2CamelS(arr[1]))
	// 				vk := rv.Kind().String()
	// 				if vk == "int64" {
	// 					return fmt.Sprintf("%d", rv.Int())
	// 				}
	// 				if vk == "string" {
	// 					return rv.String()
	// 				}
	// 				// 继续向下查找
	// 				return rv.String()
	// 			}
	// 		}
	// 	}
	// 	fmt.Println("::", tagName, thisv)
	// }

	if thisv.IsValid() {
		if arr := strings.Split(tagName, "."); len(arr) > 1 {
			tmp := *(thisv)
			for j := 0; j < len(arr); j++ {
				if tv := tmp.Elem().FieldByName(case2CamelS(arr[j])); tv.IsValid() {
					tmp = tv
					tkind := tmp.Kind().String()
					if tkind == "string" {
						return tmp.String()
					}
					if IsKindInt(tkind) {
						return fmt.Sprintf("%d", tmp.Int())
					}
				}
			}
		}
	}

	// 一次性访问两个数组是非法的
	if thisv.IsValid() && thisvKind == "slice" {
		return types.Err(1090, tagName)
	}

	fieldName := case2CamelS(tagName)
	if m := thisv.Elem().FieldByName(fieldName); m.IsValid() {
		if vk := m.Kind().String(); IsKindInt(vk) {
			return strconv.FormatInt(m.Int(), 10)
		}
		return m.String()
	}
	return types.Err(1097, tagName)
}

// case2CamelS .
func case2CamelS(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

func Camel2Case(name string) string {
	str := ""
	for i := 0; i < len(name); i++ {
		k := name[i]
		if k == '_' || i == 0 {
			str += strings.ToUpper(string(k))
		} else {
			str += string(k)
		}
	}
	return str
}

func IsKindInt(k string) bool {
	return k == "int" || k == "int64" || k == "int32" || k == "int8" || k == "int16"
}
