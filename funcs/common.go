package funcs

// IsKindInt 判断当前 reflect 类型是否为整型
func IsKindInt(k string) bool {
	return k == "int" || k == "int64" || k == "int32" || k == "int8" || k == "int16"
}

// IsKindFloat 判断当前 reflect 类型是否为整型
func IsKindFloat(k string) bool {
	return k == "float" || k == "float64" || k == "float32"
}
