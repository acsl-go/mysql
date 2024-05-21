package mysql

import "strings"

func camelToSnake(s string) string {
	var sb strings.Builder
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				sb.WriteByte('_')
			}
			sb.WriteByte(byte(c + 32))
		} else {
			sb.WriteByte(byte(c))
		}
	}
	return sb.String()
}
