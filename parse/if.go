package parse

import (
	"bytes"
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"reflect"
	"strings"
	"unicode"

	"github.com/pywee/gtpl/funcs"
	"github.com/pywee/gtpl/types"
)

// SplitIfExt 分离 if, else, else if 代码块
// 并进行表达式解析、if 优先级判断等
// 返回最终结果的最终结果必须是一块局部代码块
func SplitIfExt(src []byte, s, e int, vars []*reflect.Value, v1 *reflect.Value) ([]byte, error) {
	k := 0
	slen := len(src)
	tmp := make([]byte, 0, 40)
	m := make([][]byte, 0, 2)
	for i := 0; i < slen; i++ {
		if src[i] == '{' && i+1 < slen {
			if src[i+1] == '!' {
				k++
			} else if src[i+1] == '/' {
				k--
			} else if src[i+1] == ';' {
				m = append(m, tmp)
				tmp = nil
			}
		}
		tmp = append(tmp, src[i])
	}

	// 闭合处
	lsidx := bytes.LastIndex(tmp, []byte("{/"))
	if lsidx == -1 {
		return nil, types.Errn(1094)
	}
	m = append(m, bytes.TrimRight(tmp[:lsidx], ""))

	for _, v := range m {
		if idx := bytes.Index(v, []byte("}")); idx != -1 {
			ret, err := getIfExtenses(v[:idx])
			if err != nil {
				return nil, err
			}

			lts := make([][]byte, 0, 2)
			letter := make([]byte, 0, 5)
			for _, p := range ret {
				if unicode.IsLetter(rune(p)) || p == '_' {
					letter = append(letter, p)
				} else if len(letter) > 0 {
					lts = append(lts, letter)
					letter = nil
				}
			}
			if len(letter) > 0 {
				lts = append(lts, letter)
			}

			// 替换变量为实际的值
			for _, lt := range lts {
				for _, vs := range vars {
					if vs == nil || !vs.IsValid() {
						continue
					}
					if vs.Kind().String() == "slice" {
						vs = v1
					}
					fv := vs.Elem().FieldByName(string(case2Camel(lt)))
					if fv.IsValid() {
						if ft := fv.Type().String(); funcs.IsKindInt(ft) {
							vv := fv.Int()
							ret = bytes.Replace(ret, lt, []byte(fmt.Sprintf("%d", vv)), 1)
						} else if ft == "string" {
							vv := fv.String()
							ret = bytes.Replace(ret, lt, []byte(vv), 1)
						}
					}
				}
			}

			// else
			if ret == nil {
				return v[idx+1:], nil
			}

			result, err := parseIfExt(ret)
			if err != nil {
				return nil, err
			}
			if result == "true" {
				return v[idx+1:], nil
			}
		}
	}

	return nil, nil
}

// getIfExtenses 从提取到的 if 条件中提取表达式
// 返回最终的表达式条件语句 1+1==2
func getIfExtenses(f []byte) ([]byte, error) {
	msg := "[if 表达式错误，请检查语法是否合法。多余的空格，或关键字书写错误，都将触发这个错误，引发错误的语句：" + string(f) + "}]"

	f = bytes.TrimLeft(f, "{!")
	f = bytes.TrimLeft(f, "{;")
	f = bytes.TrimSpace(f)

	arr := bytes.Split(f, []byte{32})
	alen := len(arr)
	if alen == 0 {
		return nil, errors.New(msg)
	}

	a0 := bytes.ToLower(bytes.TrimSpace(arr[0]))
	if alen == 1 && bytes.Equal(a0, []byte("else")) {
		return nil, nil
	}

	// alen >= 1
	if bytes.Equal(a0, []byte("if")) || bytes.Equal(a0, []byte("elseif")) {
		m := bytes.Join(arr[1:], nil)
		return m, nil
	}

	return nil, errors.New(msg)
}

// ParseIfExt 解析表达式
// 返回的结果为 bool 类字符串
func parseIfExt(src []byte) (string, error) {
	var (
		s       scanner.Scanner
		k       = 0     // 用于标记括号 () 的出现次数，当它等于0时表示已经找到最里层的括号
		j       = false // j==true时提取字符串
		flot    = false // 是否存在浮点型计算
		ext     = ""    // 完整的字符串表达式 在收集过程中不断从递归中计算得出结果
		compare = ""    // 是否存在比较运算 如 "==" 符号 1 ==,  2 !=, 3 >, 4 <, 5 >=, 6 <=
		comtype = 0     // 1-包含比较运算符 2-包含&&||
		bracks  = make([]byte, 0, 10)
		fset    = token.NewFileSet()
		srcLen  = len(src)
		file    = fset.AddFile("", fset.Base(), srcLen)
	)

	s.Init(file, src, nil, scanner.ScanComments)
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		tk := tok.String()
		if tk == ";" {
			continue
		}
		if tk == "(" {
			k++
			j = true
		} else if tk == ")" {
			k--
			if k == 0 {
				xt, err := parseIfExt(bracks[1:])
				if err != nil {
					return "", err
				}
				ext += xt
				bracks = nil
				j = false
			}
		}

		if !j && tk != ")" && tk != "(" {
			switch tk {
			case "&&", "||":
				if comtype < 2 {
					comtype = 2
				}
				compare = tk
			case "==", "!=", ">", "<", ">=", "<=":
				if comtype == 0 {
					comtype = 1
				}
				compare = tk
			case "FLOAT":
				flot = true
			}
			if lit != "" {
				ext += lit
			} else {
				ext += tk
			}
		}

		if j {
			if lit != "" {
				bracks = append(bracks, []byte(lit)...)
			} else {
				bracks = append(bracks, []byte(tk)...)
			}
		}
		// fmt.Printf("%s\t%s\t%q\n", fset.Position(pos), tok, lit)
		// fmt.Printf("%s\t%q\n", tok, lit)
	}

	// 逻辑运算 "&&" 和 "||"
	if comtype == 2 {
		// fmt.Println(compareSymOr(ext))
		return compareSymOr(ext)
	}

	// 比较运算
	if comtype == 1 {
		rel, err := splitCompareSymbool(ext, compare)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", rel), nil
	}

	if flot {
		r, err := calculationFloat(ext)
		if err != nil {
			return "", err
		}
		return r, nil
	}

	r, err := calculationInt(ext)
	if err != nil {
		return "", err
	}
	return r, nil
}

// compareSymOr 逻辑运算 "||" 和 "&&"
func compareSymOr(src string) (string, error) {
	arr := strings.Split(src, "||")
	for _, v := range arr {
		if v == "false" {
			continue
		}
		final := true // 记录最终判断
		mr := strings.Split(v, "&&")
		for _, n := range mr {
			if n == "false" {
				final = false
				break
			}
			r, err := splitCompareSymbool(n, findEqualSymbool(n))
			if err != nil {
				return "false", err
			}
			if !r {
				final = false
				break
			}
		}
		if final {
			return "true", nil
		}
	}
	return "false", nil
}

func findEqualSymbool(src string) string {
	if strings.Contains(src, "==") {
		return "=="
	}
	if strings.Contains(src, "!=") {
		return "!="
	}
	if strings.Contains(src, ">=") {
		return ">="
	}
	if strings.Contains(src, "<=") {
		return "<="
	}
	if strings.Contains(src, ">") {
		return ">"
	}
	if strings.Contains(src, "<") {
		return "<"
	}
	return ""
}

// splitCompareSymbool 比较运算
// == != >= <= > <
func splitCompareSymbool(src, compare string) (bool, error) {
	arr := strings.Split(src, compare)
	arr0 := arr[0]
	arr1 := arr[1]

	b0, b0s := checkStringType(arr0)
	if b0 == 0 {
		nb, err := parseIfExt([]byte(arr0))
		if err != nil {
			return false, err
		}
		b0, b0s = checkStringType(nb)
	}

	b1, b1s := checkStringType(arr1)
	if b1 == 0 {
		nb, err := parseIfExt([]byte(arr1))
		if err != nil {
			return false, err
		}
		b1, b1s = checkStringType(nb)
	}

	// 布尔值比较
	if (b0 == 4 || b0 == 5) && (b1 == 4 || b1 == 5) {
		if compare == "==" {
			return b0s == b1s, nil
		}
		if compare == "!=" {
			return b0s != b1s, nil
		}
		return false, types.Errn(1096)
	}

	// 字符串比较 两者中如果只有一个是字符串
	// 则不允许参与比较
	if (b0 == 3 && b1 != 3) || (b1 == 3 && b0 != 3) {
		return false, errors.New(types.StrCanNotBeCompared)
	}
	if b0 == 3 {
		if compare == "==" {
			return b0s == b1s, nil
		}
		if compare == "!=" {
			return b0s != b1s, nil
		}
		return false, types.Errn(1096)
	}
	if b0 == 2 || b1 == 2 {
		return compareFloat(arr0, arr1, compare)
	}
	return compareInt(arr0, arr1, compare)
}

// parseIf .
func parseIf(tagName []byte, vv []*reflect.Value, values *reflect.Value) (bool, error) {
	t, err := parseIfExt(tagName)
	if err != nil {
		return false, err
	}

	if t == "false" {
		return false, nil
	}

	if t == "true" {
		return true, nil
	}
	return false, types.Errn(1095)
}

// parseString 语法检查
func parseString(s []byte) (string, bool) {
	slen := len(s)
	s92 := bytes.Count(s, []byte{34})
	s34 := bytes.Count(s, []byte{92})
	if sr := s92 + s34; sr > 0 && sr%2 != 0 {
		return "", false
	}

	if slen > 1 {
		qs := s[1 : slen-1]
		for j := 0; j < len(qs); j++ {
			if qs[j] == 34 && j-1 >= 0 && qs[j-1] != '\\' {
				return "", false
			}
		}
	}

	s = bytes.Replace(s, []byte{92, 34}, []byte{34}, -1)
	s = bytes.Replace(s, []byte{92, 92}, []byte{92}, -1)
	if s[0] == 34 && s[len(s)-1] == 34 {
		s = s[1 : len(s)-1]
	}
	return string(s), true
}

// case2Camel .
func case2Camel(name []byte) []byte {
	name = bytes.Replace(name, []byte("_"), []byte{32}, -1)
	name = bytes.Title(name)
	return bytes.Replace(name, []byte{32}, nil, -1)
}

// strings.NewReplacer(
// 	"+", " ",
// 	"-", " ",
// 	"*", " ",
// 	"/", " ",
// 	"&&", " ",
// 	"||", " ",
// 	"==", " ",
// 	">=", " ",
// 	"<=", " ",
// 	">", " ",
// 	"<", " ",
// 	"^", " ",
// )
