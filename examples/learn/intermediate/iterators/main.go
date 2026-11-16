// ตัวอย่าง iterators (range over func) ตั้งแต่ Go 1.23
package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

// Countdown คืน iter.Seq[int]: นับถอยหลังจาก n ถึง 1
func Countdown(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := n; i > 0; i-- {
			if !yield(i) { // ผู้ใช้ break แล้ว → หยุดทันที
				return
			}
		}
	}
}

// Fibonacci เป็น iterator ไม่มีที่สิ้นสุด ผู้เรียกต้อง break เอง
func Fibonacci() iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1
		for {
			if !yield(a) {
				return
			}
			a, b = b, a+b
		}
	}
}

// Filter รับ Seq แล้วคืน Seq ใหม่ที่ผ่านเงื่อนไข (ต่อกันเป็น pipeline ได้)
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Lines คืน iter.Seq2 คู่ (บรรทัดที่, ข้อความ)
func Lines(s string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, line := range strings.Split(s, "\n") {
			if !yield(i+1, line) {
				return
			}
		}
	}
}

func main() {
	for n := range Countdown(3) {
		fmt.Print(n, " ")
	}
	fmt.Println("go!")

	// break ได้ตามปกติ iterator จะได้ false จาก yield
	for f := range Fibonacci() {
		if f > 50 {
			break
		}
		fmt.Print(f, " ")
	}
	fmt.Println()

	even := Filter(Fibonacci(), func(n int) bool { return n%2 == 0 })
	for v := range even {
		if v > 1000 {
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()

	for no, line := range Lines("alpha\nbeta\ngamma") {
		fmt.Printf("%d: %s\n", no, line)
	}

	// ฟังก์ชันใน slices และ maps ที่คืน iterator
	fruits := []string{"mango", "durian", "banana"}
	for i, f := range slices.All(fruits) {
		fmt.Println(i, f)
	}
	long := slices.Collect(Filter(slices.Values(fruits), func(s string) bool {
		return len(s) > 5
	}))
	fmt.Println(long)

	stock := map[string]int{"mango": 5, "durian": 2, "banana": 12}
	keys := slices.Sorted(maps.Keys(stock)) // map ไม่มีลำดับ → sort ก่อนแสดง
	fmt.Println(keys)
}
