package templates

import (
	"crypto/md5"
	"fmt"
	"gtpl/types"
	"reflect"
	"strconv"
	"strings"
)

type fns struct {
	id     string
	fname  string
	params []string
	Result string
}

// findFuncs 重要函数
// 找出当前标签中的函数表达式 并递归将其线性输出 使执行过程变为线性
func findFuncs(tagName string, i int) ([]*fns, error) {
	i++
	var f = make([]*fns, 0, 3)
	if idx := strings.LastIndex(tagName, "("); idx != -1 {
		if strings.Count(tagName, "(") != strings.Count(tagName, ")") {
			return nil, types.Errn(1095, tagName)
		}
		funcName := new(fns)
		idx2 := -1
		if idx2 = strings.LastIndex(tagName[:idx], ","); idx2 != -1 {
			funcName.fname = strings.TrimSpace(tagName[idx2+1 : idx])
		} else if idx2 = strings.LastIndex(tagName[:idx], "("); idx2 != -1 {
			funcName.fname = strings.TrimSpace(tagName[idx2+1 : idx])
		} else {
			funcName.fname = tagName[:idx]
		}
		tmp := tagName[idx+1:]
		if idx3 := strings.Index(tmp, ")"); idx3 != -1 {
			funcName.id = "_key:" + strconv.Itoa(i)
			funcName.params = strings.Split(tmp[:idx3], ",")
			tagName = strings.Replace(tagName, tagName[idx2+1:idx]+tagName[idx:idx+idx3+2], funcName.id, 1)
			f = append(f, funcName)
			m, err := findFuncs(tagName, i)
			if err != nil {
				return nil, err
			}
			f = append(f, m...)
		}
	}
	return f, nil
}

// CheckFuncs 查找是否有调用函数
// 如果有，返回函数列表
func CheckFuncs(tagName string) ([]string, string) {
	flist := make([]string, 0, 2)
	tlen := len(tagName)
	if idx := strings.Index(tagName, "("); idx != -1 {
		if tagName[tlen-1] != 41 {
			return flist, tagName
		}
		flist = append(flist, tagName[:idx])
		tlist, tname := CheckFuncs(tagName[idx+1 : tlen-1])
		flist = append(flist, tlist...)
		tagName = tname
	}
	return flist, tagName
}

// execFunc 解析并执行函数
// trim ltrim rtrim trimspace tolower toupper replace md5 len
func execFunc(f *fns, vv []*reflect.Value) string {
	plen := len(f.params)
	for i := 0; i < plen; i++ {
		j := strings.Trim(f.params[i], " ")
		jlen := len(j)
		if jlen > 1 && j[0] == '"' && j[jlen-1] == '"' {
			j = j[1 : jlen-1]
		}
		f.params[i] = j
	}

	// 当前结果是否包含上一次变量计算结果
	p0 := f.params[0]
	fundResult := false
	if len(p0) > 5 && p0[:5] == "_key:" {
		fundResult = true
	}

	vlen := len(vv)
	switch strings.ToLower(f.fname) {
	case "len":
		if !fundResult {
			return strconv.Itoa(strings.Count(parseVal(p0, vv, vv[vlen-1]), "") - 1)
		}
		return strconv.Itoa(strings.Count(f.Result, "") - 1)
	case "tolower":
		if !fundResult {
			return strings.ToLower(parseVal(p0, vv, vv[vlen-1]))
		}
		return strings.ToLower(f.Result)
	case "toupper":
		if !fundResult {
			return strings.ToUpper(parseVal(p0, vv, vv[vlen-1]))
		}
		return strings.ToUpper(f.Result)
	case "md5":
		val := f.Result
		if !fundResult {
			val = parseVal(p0, vv, vv[vlen-1])
			if len(val) > 2 && val[:2] == "{:" {
				val = p0
			}
		}
		return fmt.Sprintf("%x", md5.Sum([]byte(val)))
	case "trim":
		p1 := " "
		if plen > 1 {
			p1 = f.params[1]
		}
		if !fundResult {
			return strings.Trim(parseVal(p0, vv, vv[vlen-1]), p1)
		}
		return strings.Trim(f.Result, p1)
	case "ltrim":
		p1 := " "
		if plen > 1 {
			p1 = f.params[1]
		}
		if !fundResult {
			return strings.TrimLeft(parseVal(p0, vv, vv[vlen-1]), p1)
		}
		return strings.TrimLeft(f.Result, p1)
	case "rtrim":
		p1 := " "
		if plen > 1 {
			p1 = f.params[1]
		}
		if !fundResult {
			return strings.TrimRight(parseVal(p0, vv, vv[vlen-1]), p1)
		}
		return strings.TrimRight(f.Result, p1)
	case "trimspace":
		if !fundResult {
			return strings.TrimSpace(parseVal(p0, vv, vv[vlen-1]))
		}
		return strings.TrimSpace(f.Result)
	case "replace":
		n := -1
		if plen > 3 {
			n, _ = strconv.Atoi(f.params[3])
		} else if plen < 3 {
			return ""
		}
		if !fundResult {
			return strings.Replace(parseVal(p0, vv, vv[vlen-1]), f.params[1], f.params[2], n)
		}
		return strings.Replace(f.Result, f.params[1], f.params[2], n)
	}
	return ""
}
