// function: multiple return, named results, variadic, first-class function, closure
package main

import (
	"errors"
	"fmt"
	"strings"
)

// function พื้นฐาน: parameter ที่ type เหมือนกันเขียนรวบได้
func add(a, b int) int {
	return a + b
}

// คืนค่าหลายค่า: รูปแบบที่พบบ่อยที่สุดคือ (ผลลัพธ์, error)
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("หารด้วยศูนย์ไม่ได้")
	}
	return a / b, nil
}

// named results: ตั้งชื่อค่าที่คืน ทำหน้าที่เป็นตัวแปรใน function
func minMax(nums []int) (lo, hi int) {
	lo, hi = nums[0], nums[0]
	for _, n := range nums[1:] {
		lo = min(lo, n)
		hi = max(hi, n)
	}
	return lo, hi
}

// variadic: รับ argument กี่ตัวก็ได้ ภายใน function nums คือ []int
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// function type: ตั้งชื่อให้ "รูปร่าง" ของ function
type transform func(string) string

// รับ function เป็น parameter (higher-order function)
func applyAll(s string, fns ...transform) string {
	for _, fn := range fns {
		s = fn(s)
	}
	return s
}

// คืน function ออกไป: closure จับตัวแปร count ไว้
func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	fmt.Println(add(2, 3))

	if q, err := divide(10, 4); err == nil {
		fmt.Println("10 / 4 =", q)
	}
	if _, err := divide(1, 0); err != nil {
		fmt.Println("error:", err)
	}

	lo, hi := minMax([]int{7, 2, 9, 4})
	fmt.Println("min:", lo, "max:", hi)

	fmt.Println(sum(), sum(1, 2), sum(1, 2, 3, 4))
	scores := []int{10, 20, 30}
	fmt.Println(sum(scores...)) // กระจาย slice เป็น argument

	// function เป็นค่า: เก็บใส่ตัวแปรได้
	shout := func(s string) string { return strings.ToUpper(s) + "!" }
	fmt.Println(applyAll("  hello go  ", strings.TrimSpace, shout))

	// closure: แต่ละ counter มี count เป็นของตัวเอง
	next := counter()
	fmt.Println(next(), next(), next())
	other := counter()
	fmt.Println(other())

	// defer: เลื่อนการทำงานไปตอน function จบ (ทำแบบ LIFO)
	for i := range 3 {
		defer fmt.Println("defer", i)
	}
	fmt.Println("จบ main")
}
