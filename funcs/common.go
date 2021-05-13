package funcs

import (
	"bytes"
	"strings"
)

// IsKindInt 判断当前 reflect 类型是否为整型
func IsKindInt(k string) bool {
	return k == "int" || k == "int64" || k == "int32" || k == "int8" || k == "int16"
}

// IsKindFloat 判断当前 reflect 类型是否为整型
func IsKindFloat(k string) bool {
	return k == "float" || k == "float64" || k == "float32"
}

// Case2Camel 小转大
func Case2Camel(name []byte) []byte {
	name = bytes.Replace(name, []byte("_"), []byte{32}, -1)
	name = bytes.Title(name)
	return bytes.Replace(name, []byte{32}, nil, -1)
}

// Case2CamelS 小转大
func Case2CamelS(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}
