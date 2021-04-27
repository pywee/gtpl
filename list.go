package gtpl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"gtpl/types"
)

type Parser struct {
	src []byte
}

func NewParser() *Parser {
	return new(Parser)
}

// Fprint 输出数据到屏幕
func (p *Parser) Fprint(i io.Writer) (n int, err error) {
	return fmt.Fprintf(i, "%s", bytes.Trim(p.src, "\r\n"))
}

func (p *Parser) String() string {
	if p == nil {
		return ""
	}
	return string(bytes.Trim(p.src, "\r\n"))
}

// ParseFile 通过文件解析数据
func (p *Parser) ParseFile(file string, data interface{}) (*Parser, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return p.Parse(src, data)
}

// Parse 通过 byte 解析数据
func (p *Parser) Parse(file []byte, data interface{}) (*Parser, error) {
	src := []byte{13, 10}
	src = append(src, file...)
	src = append(src, []byte{13, 10}...)

	dl := reflect.ValueOf(data)
	rel, err := Pe(src, nil, &dl)
	if err != nil {
		return nil, err
	}
	p.src = bytes.TrimRight(rel, "\n ")
	return p, nil
}

// Pe .
func Pe(src []byte, vars []*reflect.Value, values *reflect.Value) ([]byte, error) {
	var (
		k   = 0
		s   = 0
		e   = 0
		sl  = len(src)
		str = make([]byte, 0, sl)
	)

	if values != nil {
		vars = append(vars, values)
	}
	for i := 0; i < sl; i++ {
		i1 := i + 1
		if src[i] == '{' && i1 < sl && isTagAllow(src[i1]) {
			si1 := src[i1]
			if si1 == '!' { // 标签列表开始前缀
				k++
				if s == 0 {
					s = i
				}
			} else if si1 == '/' { // 闭合标签
				k--
			} else if s == 0 && e == 0 {
				if si1 == '.' { // 通用全局变量调用
					n := getNextI(src, sl, i)
					// str = append(str, parseGlobalVar(src[i+2:n])...)
					str = append(str, parseGlobalStructVal(string(src[i+2:n]), values)...)
					i = n
					continue
				}
				if si1 == ':' { // 作用域变量调用
					n := getNextI(src, sl, i)
					srcTagName := src[i+2 : n]
					if jlist, err := findFuncs(string(srcTagName), 0); err != nil {
						return nil, err
					} else if jlen := len(jlist); jlen > 0 {
						r := ""
						for m := 0; m < jlen; m++ {
							r = execFunc(jlist[m], vars)
							if m+1 < len(jlist) {
								jlist[m+1].Result = r
							}
						}
						str = append(str, []byte(r)...)
					} else {
						buf := parseVal(string(srcTagName), vars, values)
						str = append(str, []byte(buf)...)
					}
					i = n
					continue
				}
			}
			if k == 0 {
				for j := i; j+1 < sl; j++ {
					if src[j] == '}' {
						e = j + 1
						i = e
						break
					}
				}
				data, err := parseTags(src[s:e], vars, values)
				if slen := len(str); slen > 0 && len(bytes.TrimSpace(str)) == 0 {
					str = nil
				}
				if err != nil {
					return nil, err
				}
				str = append(str, data...)
				s = 0
				e = 0
			}
		}

		// 普通字符串
		if s == 0 && e == 0 {
			str = append(str, src[i])
		}
	}

	return str, nil
}

// parseTags 解析标签局部代码块
// 拨开最外层标签，然后递归处理标签内部代码块数据
func parseTags(src []byte, vars []*reflect.Value, v1 *reflect.Value) ([]byte, error) {
	ts := bytes.Index(src, []byte("}"))
	te := bytes.LastIndex(src, []byte("{"))
	if ts == -1 || te == -1 {
		return src, types.Errn(1093)
	}

	// 提取最外层标签
	tinfo, err := getTagInfo(src[:ts+1], src[te:])
	if err != nil {
		return nil, err
	}

	// 处理 if 语句块
	// 包含了 if,elseif,else 三类归为一整块
	if tinfo.name == "if" {
		rel, err := SplitIfExt(src, ts, te, vars, v1)
		if err != nil {
			return nil, err
		}

		// fmt.Println(string(rel))

		return Pe(rel, vars, v1)
	}

	// 提取了最外层标签后剩下的局部数据
	theRestStr := src[ts+1 : te]
	strParsed := make([]byte, 0, 100)

	// 查找结构体
	// resStruct, err := parseReflectSlice(vars, theRestStr)
	// if err == nil && len(resStruct) > 0 {
	// 	return resStruct, nil
	// }

	resStruct, err := parseReflectSlice(vars, string(tinfo.model), theRestStr)
	if err != nil {
		return nil, err
	}
	if len(resStruct) > 0 {
		return resStruct, nil
	}
	return strParsed, nil
}

// parseReflectSlice 解析实际数据而非 []*reflect.Values
func parseReflectSlice(vars []*reflect.Value, fieldName string, theRestStr []byte) ([]byte, error) {
	strParsed := make([]byte, 0, 100)
	vs := vars[len(vars)-1]

	if !vs.IsValid() {
		return nil, nil
	}

	structName := vs.Type().Name()
	if len(structName) == 0 {
		structName = vs.Type().String()
	}

	if idx := strings.Index(structName, "."); idx != -1 {
		if rs := structName[idx+1:]; case2CamelS(fieldName) != rs {
			arr := strings.Split(fieldName, ".")
			alen := len(arr)
			if alen > 2 {
				return nil, types.Errn(1092)
			}

			// if alen <= 1 || case2CamelS(arr[0]) != rs {
			// 	return nil, errors.New("找不到要输出的数据：" + fieldName)
			// }

			if alen > 1 && case2CamelS(arr[0]) == rs {
				if vs.Kind().String() == "slice" {
					return nil, types.Errn(1091, structName)
				}
				field := vs.Elem().FieldByName(case2CamelS(arr[1]))
				vs = &field
			}
		}
	}

	if vs.Type().Kind().String() == "slice" {
		for j := 0; j < vs.Len(); j++ {
			rv := vs.Index(j)
			rel, err := Pe(theRestStr, vars, &rv)
			if err != nil {
				return nil, err
			}
			rel = bytes.TrimRight(rel, "\n ") // 关键
			strParsed = append(strParsed, rel...)
		}
		return strParsed, nil
	}

	// ptr
	// 解析结构体内的数组字段
	field := vs.Elem().FieldByName(case2CamelS(fieldName))
	vars = append(vars, &field)
	rel, err := parseReflectSlice(vars, fieldName, theRestStr)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

// getNextI .
func getNextI(detail []byte, dlen, i int) int {
	for j := i; j < dlen; j++ {
		if detail[j] == '}' {
			return j
		}
	}
	return 0
}

// allowTags 允许的tag前缀 需要组合{字符
var allowTags = []byte{'!', ':', '/', '.'}

// isTagAllow 检查当前tag前缀是否允许
func isTagAllow(tag byte) bool {
	for _, t := range allowTags {
		if t == tag {
			return true
		}
	}
	return false
}
