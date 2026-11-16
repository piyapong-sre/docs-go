package calc_test

import (
	"fmt"

	calc "learn/intermediate/testing" // โฟลเดอร์ชื่อ testing แต่ package ชื่อ calc
)

// Example test: เป็นทั้งเอกสาร (โชว์ใน go doc) และ test (เทียบกับ // Output:)
func ExampleAdd() {
	fmt.Println(calc.Add(1, 2))
	// Output: 3
}

func ExampleDivide() {
	_, err := calc.Divide(1, 0)
	fmt.Println(err)
	// Output: divide by zero
}
