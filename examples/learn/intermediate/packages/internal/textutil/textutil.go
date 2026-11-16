// Package textutil เป็น internal package:
// import ได้เฉพาะ code ที่อยู่ใต้ learn/intermediate/packages/ เท่านั้น
package textutil

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Title ทำให้ตัวอักษรแรกเป็นตัวพิมพ์ใหญ่
func Title(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
}

// Reverse กลับลำดับตัวอักษร (ทำงานระดับ rune จึงรองรับ UTF-8)
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
