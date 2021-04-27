package template

import (
	"bytes"
	"errors"
)

// srcInfo 解析标签头部之后得到的数据
// 如 {!list:article;cid=1}
type srcInfo struct {
	name   string // list, sql, if 等
	model  []byte // cate, article
	params [][]byte
}

// getTagInfo 获取tag信息，如 list, sql, if 等
func getTagInfo(f, e []byte) (*srcInfo, error) {
	result := new(srcInfo)
	f = bytes.TrimLeft(f, "{!")
	f = bytes.TrimRight(f, "}")
	e = bytes.TrimRight(e[2:], "}")
	e = bytes.TrimSpace(e)
	if idx := bytes.Index(f, []byte(":")); idx != -1 {
		result.name = string(bytes.Trim(f[:idx], " "))
		f = f[idx+1:]
		arr := bytes.Split(f, []byte(";"))
		if alen := len(arr); alen > 0 {
			if result.name != string(e) {
				return nil, errors.New("闭合标签有误，请检查:{/" + string(e) + "}")
			}
			if result.name == "list" {
				result.model = arr[0]
				if alen > 1 {
					pars := make([][]byte, 0, alen-1)
					for _, v := range arr[1:] {
						pars = append(pars, bytes.TrimSpace(v))
					}
					result.params = pars
				}
			} else if result.name == "sql" {
				result.model = bytes.TrimSpace(f)
			}
		}
	} else if arr := bytes.Split(f, []byte{32}); len(arr) > 1 {

		// todo 判断闭合标签名称是否与开头相符
		// ...

		result.name = string(bytes.TrimSpace(arr[0]))
		result.model = arr[1]
	}
	return result, nil
}
