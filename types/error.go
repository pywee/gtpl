package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	InvalidExtention          = "无效表达式 "
	InvalidCompareExts        = "无效的比较运算"
	NotFoundValid             = "找不到变量"
	ErrorIfExtention          = "if 语句错误，找不到对称的标签闭合语法"
	IllegalCrossTheBorder     = "非法定义，标签解析越界"
	StrCanNotBeCompared       = "字符串不可与其他类型进行比较"
	ExtentionNotAllowToRange  = "非法操作: 不可以一次性循环三级数据"
	ExtentionNotAllowForSlice = "无效的访问方式，如果最外层是数组，不可直接使用链式操作"
	ExtentionNotAllowToUse    = "错误的访问方式，结构体最外层已经是数组，不可直接使用链式操作:"
)

func Err(code int16, msg ...string) string {
	var f string

	switch code {
	case 1090:
		f = ExtentionNotAllowForSlice
	case 1091:
		f = ExtentionNotAllowToUse
	case 1092:
		f = ExtentionNotAllowToRange
	case 1093:
		f = IllegalCrossTheBorder
	case 1094:
		f = ErrorIfExtention
	case 1095:
		f = InvalidExtention
	case 1096:
		f = InvalidCompareExts
	case 1097:
		f = NotFoundValid
	}

	if len(msg) > 0 {
		f += " " + strings.Join(msg, "")
	}
	return fmt.Sprintf("[error code: %d; msg: %s]", code, f)
}

func Errn(code int16, msg ...string) error {
	return errors.New(Err(code, msg...))
}
