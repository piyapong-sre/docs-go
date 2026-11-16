// Package calc เป็นตัวอย่าง package เล็กๆ สำหรับฝึกเขียน test
package calc

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrDivideByZero = errors.New("divide by zero")

// Add บวกเลขสองตัว
func Add(a, b int) int { return a + b }

// Divide หารเลข คืน error ถ้าตัวหารเป็นศูนย์
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

// Reverse กลับลำดับตัวอักษรแบบรองรับ UTF-8
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// WordCount นับจำนวนคำ (คั่นด้วย whitespace)
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// IsValidUTF8 ช่วยตรวจ input ใน fuzz test
func IsValidUTF8(s string) bool { return utf8.ValidString(s) }
