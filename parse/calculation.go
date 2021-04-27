package parse

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/pywee/gtpl/types"

	"github.com/shopspring/decimal"
)

// checkStringType 取得字符串类型
// 0-无法识别 1-整型 2-浮点型 3-字符串 4-布尔型(true) 5-布尔型(false)
func checkStringType(src string) (int8, string) {
	var (
		l = len(src)
	)
	if l >= 4 {
		sl := strings.ToLower(src)
		if sl == "true" {
			return 4, "true"
		}
		if sl == "false" {
			return 5, "false"
		}
	}

	// 字符串判断
	str, err := checkAndGetString(src)
	if err != nil {
		return 0, err.Error()
	}
	if str != "" {
		return 3, str
	}

	// 数值判断
	i := isDigit(src)
	if i == 1 {
		return 1, src
	}
	if i == 2 {
		return 2, src
	}

	return 0, ""
}

// isBool .
func isBool(src string) bool {
	if src == "true" {
		return true
	}
	return src == "false"
}

// isDigit 判断字符串是否为数字字符串
// f: 0-其他类型，如字符串 1-整型 2-浮点
func isDigit(str string) int8 {
	var f int8 = 1
	for _, x := range str {
		if x == '.' {
			f = 2
			continue
		}
		if !unicode.IsDigit(x) {
			return 0
		}
	}
	return f
}

// parseBrackets 根据表达式中的括号解析 分离数据
// !已替代
func parseBrackets(txt string) []string {
	flist := make([]string, 0, 12)
	if idx := strings.Index(txt, "("); idx != -1 {
		flist = append(flist, strings.TrimSpace(txt[:idx]))
		txt = txt[idx:]
		k := 0
		for j := 0; j < len(txt); j++ {
			if txt[j] == '(' {
				k++
			} else if txt[j] == ')' {
				k--
			}
			if k == 0 {
				// fmt.Println(k, string(txt[1:j]), string(txt[j+1:]))
				flist = append(flist, parseBrackets(txt[1:j])...)
				txt = txt[j+1:]
				flist = append(flist, parseBrackets(txt)...)
				break
			}
		}
	} else {
		flist = append(flist, strings.TrimSpace(txt))
	}
	return flist
}

type di struct {
	e    string
	c    rune
	pass bool
}

// calculationInt .
// + - * / % 计算
func calculationInt(s string) (string, error) {
	var us = make([]*di, 0, 4)
	var u = &di{}
	for _, v := range s {
		if v == '+' || v == '-' || v == '*' || v == '/' {
			u.c = v
			us = append(us, u)
			u = &di{}
		} else {
			u.e += strings.TrimSpace(string(v))
		}
	}

	us = append(us, u)
	if ulen := len(us); ulen > 0 {
		for i := 0; i < ulen; i++ {
			if i == 0 {
				continue
			}
			if us[i-1].c == '*' {
				l, _ := strconv.Atoi(us[i-1].e)
				t, _ := strconv.Atoi(us[i].e)
				us[i].e = strconv.Itoa(l * t)
				us[i-1].pass = true
			} else if us[i-1].c == '/' {
				l, _ := strconv.Atoi(us[i-1].e)
				t, _ := strconv.Atoi(us[i].e)
				us[i].e = strconv.Itoa(l / t)
				us[i-1].pass = true
			}
		}
	}

	num := 0
	h := false
	l := &di{}
	for _, v := range us {
		if v.pass {
			continue
		}
		if isDigit(v.e) == 0 {
			return "", types.Errn(1095, s)
		}
		n, _ := strconv.Atoi(v.e)
		if !h {
			num = n
			h = true
		} else if l.c == '+' {
			num += n
		} else if l.c == '-' {
			num -= n
		} else if l.c == '*' {
			num *= n
		} else if l.c == '/' {
			num /= n
		}
		l = v
	}
	return strconv.Itoa(num), nil
}

// calculationFloat .
// + - * / % 计算
func calculationFloat(s string) (string, error) {
	var us = make([]*di, 0, 5)
	var u = &di{}
	for _, v := range s {
		if v == '+' || v == '-' || v == '*' || v == '/' {
			u.c = v
			us = append(us, u)
			u = &di{}
		} else {
			u.e += string(v)
			// u.e += strings.TrimSpace(string(v))
		}
	}

	us = append(us, u)
	if ulen := len(us); ulen > 0 {
		for i := 0; i < ulen; i++ {
			if i == 0 {
				continue
			}
			if isDigit(us[i].e) == 0 {
				return "", types.Errn(1095, s)
			}
			if us[i-1].c == '*' {
				l, _ := strconv.ParseFloat(us[i-1].e, 64)
				t, _ := strconv.ParseFloat(us[i].e, 64)
				us[i].e = strconv.FormatFloat(l*t, 'f', -1, 64)
				us[i-1].pass = true
			} else if us[i-1].c == '/' {
				l, _ := strconv.ParseFloat(us[i-1].e, 64)
				t, _ := strconv.ParseFloat(us[i].e, 64)
				us[i].e = strconv.FormatFloat(l/t, 'f', -1, 64)
				us[i-1].pass = true
			}
		}
	}

	var (
		h    = false
		l    = &di{}
		numf = decimal.NewFromFloat(0)
	)
	for _, v := range us {
		if v.pass {
			continue
		}
		if isDigit(v.e) == 0 {
			return "", errors.New("运算表达式非法" + s)
		}
		n, _ := strconv.ParseFloat(v.e, 64)
		nf := decimal.NewFromFloat(n)
		if !h {
			h = true
			numf = numf.Add(nf)
		} else if l.c == '+' {
			numf = numf.Add(nf)
		} else if l.c == '-' {
			numf = numf.Sub(nf)
		} else if l.c == '*' {
			numf = numf.Mul(nf)
		} else if l.c == '/' {
			numf = numf.Div(nf)
		}
		l = v
	}

	// fmt.Println(numf.String())

	return numf.String(), nil
}

// compareInt 比较整型
func compareInt(arr0, arr1, compare string) (bool, error) {
	a0, err := calculationInt(arr0)
	if err != nil {
		return false, err
	}

	a1, err := calculationInt(arr1)
	if err != nil {
		return false, err
	}

	// 数值比较
	i0, _ := strconv.Atoi(a0)
	i1, _ := strconv.Atoi(a1)
	if compare == "==" {
		return i0 == i1, nil
	}
	if compare == "!=" {
		return i0 != i1, nil
	}
	if compare == ">" {
		return i0 > i1, nil
	}
	if compare == "<" {
		return i0 < i1, nil
	}
	if compare == ">=" {
		return i0 >= i1, nil
	}
	if compare == "<=" {
		return i0 <= i1, nil
	}
	return false, errors.New("表达式不合法")
}

// compareFloat 比较浮点型
func compareFloat(arr0, arr1, compare string) (bool, error) {
	a0, err := calculationFloat(arr0)
	if err != nil {
		return false, err
	}

	a1, err := calculationFloat(arr1)
	if err != nil {
		return false, err
	}

	// 数值比较
	i0, _ := strconv.ParseFloat(a0, 64)
	i1, _ := strconv.ParseFloat(a1, 64)
	if compare == "==" {
		return i0 == i1, nil
	}
	if compare == "!=" {
		return i0 != i1, nil
	}
	if compare == ">" {
		return i0 > i1, nil
	}
	if compare == "<" {
		return i0 < i1, nil
	}
	if compare == ">=" {
		return i0 >= i1, nil
	}
	if compare == "<=" {
		return i0 <= i1, nil
	}
	return false, errors.New("表达式不合法")
}
