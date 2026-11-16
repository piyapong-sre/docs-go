// ตัวอย่าง error handling: sentinel error, custom error, wrapping, errors.Is/As/Join
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
)

// ErrNotFound เป็น sentinel error: ค่า error ที่ประกาศไว้ให้คนอื่นเปรียบเทียบได้
var ErrNotFound = errors.New("not found")

// ValidationError เป็น custom error type ที่เก็บข้อมูลเพิ่มได้
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Msg
}

var users = map[int]string{1: "somchai", 2: "somsri"}

func findUser(id int) (string, error) {
	name, ok := users[id]
	if !ok {
		// %w ห่อ error เดิมไว้ พร้อมเพิ่ม context
		return "", fmt.Errorf("find user %d: %w", id, ErrNotFound)
	}
	return name, nil
}

func parseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse age %q: %w", s, err)
	}
	if n < 0 || n > 150 {
		return 0, &ValidationError{Field: "age", Msg: "must be between 0 and 150"}
	}
	return n, nil
}

func validate(name string, age int) error {
	var errs []error
	if name == "" {
		errs = append(errs, &ValidationError{Field: "name", Msg: "required"})
	}
	if age < 18 {
		errs = append(errs, &ValidationError{Field: "age", Msg: "must be at least 18"})
	}
	return errors.Join(errs...) // คืน nil ถ้า errs ว่าง
}

func main() {
	// 1) sentinel error + errors.Is (เช็คได้แม้ถูกห่อไว้หลายชั้น)
	if _, err := findUser(99); err != nil {
		fmt.Println("error:", err)
		fmt.Println("is ErrNotFound?", errors.Is(err, ErrNotFound))
	}

	// 2) custom error type + errors.As
	if _, err := parseAge("200"); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Println("validation failed on field:", ve.Field)
		}
	}

	// 3) error จาก standard library ก็ห่อแล้วเช็คได้เหมือนกัน
	if _, err := parseAge("abc"); err != nil {
		fmt.Println("error:", err)
		fmt.Println("is strconv.ErrSyntax?", errors.Is(err, strconv.ErrSyntax))
	}

	_, err := os.Open("no-such-file.txt")
	fmt.Println("is fs.ErrNotExist?", errors.Is(err, fs.ErrNotExist))

	// 4) errors.Join รวมหลาย error เป็นก้อนเดียว
	if err := validate("", 15); err != nil {
		fmt.Println("validate:")
		fmt.Println(err)
	}
	fmt.Println("valid input ->", validate("somchai", 30))
}
