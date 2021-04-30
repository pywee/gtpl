package parse

import (
	"errors"
	"strings"
)

// checkAndGetString 检查并返回字符串
func checkAndGetString(src string) (string, bool, error) {
	if slen := len(src); slen > 0 {
		if src[0] == '"' && src[slen-1] == '"' {
			if strings.Contains(src, "\"") || strings.Contains(src, "\\") {
				yc := strings.Count(src, "\"")
				xc := strings.Count(src, "\\")
				if c := yc + xc; c%2 != 0 {
					return "", false, errors.New("字符串格式不正确")
				}
				if src[0] == '"' && src[slen-1] == '"' {
					src = src[1 : slen-1]
				}
				src = strings.Replace(src, "\\\"", "\"", -1)
				src = strings.Replace(src, "\\\\", "\\", -1)
			}
			return src, true, nil
		}
	}

	// 深入判断是否为字符串
	for _, v := range src {
		if (v < 40 || v > 58) && v != 37 { // 没有将运算符 ^ 计算在内
			return src, true, nil
		}
	}

	return src, false, nil
}
