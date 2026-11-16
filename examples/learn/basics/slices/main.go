// array และ slice
package main

import (
	"fmt"
	"slices"
)

func main() {
	// array: ขนาดเป็นส่วนหนึ่งของ type และ copy ทั้งก้อนเมื่อ assign
	a := [3]int{1, 2, 3}
	b := a // copy
	b[0] = 100
	fmt.Println("array:", a, b, len(a))

	// slice: มุมมอง (view) ไปยัง array ที่อยู่เบื้องหลัง
	s := []int{10, 20, 30, 40, 50}
	fmt.Println("slice:", s, "len:", len(s), "cap:", cap(s))

	// make([]T, len, cap)
	m := make([]int, 0, 3)
	fmt.Println("make:", m, len(m), cap(m))

	// append: ถ้า cap ไม่พอจะจอง array ใหม่ที่ใหญ่ขึ้น
	var grow []int
	prevCap := cap(grow)
	for i := range 10 {
		grow = append(grow, i)
		if cap(grow) != prevCap {
			fmt.Printf("len=%-2d cap %d -> %d\n", len(grow), prevCap, cap(grow))
			prevCap = cap(grow)
		}
	}

	// slicing: s[low:high] ได้ช่วง low..high-1 และ "แชร์" array เดียวกัน
	part := s[1:3]
	fmt.Println("part:", part, "len:", len(part), "cap:", cap(part))
	part[0] = 999 // แก้ part กระทบ s ด้วย!
	fmt.Println("s after part[0]=999:", s)

	// append บน sub-slice ที่ยังมี cap เหลือ จะไปทับข้อมูลของ s
	part = append(part, 777)
	fmt.Println("s after append(part):", s)

	// full slice expression s[low:high:max] จำกัด cap เพื่อกันการทับ
	safe := s[1:3:3]
	safe = append(safe, 555) // cap ไม่พอ -> จอง array ใหม่
	fmt.Println("s after append(safe):", s, "safe:", safe)

	// copy: คัดลอกข้อมูลจริง ได้ slice ที่เป็นอิสระ
	dst := make([]int, len(s))
	n := copy(dst, s)
	dst[0] = -1
	fmt.Println("copied", n, "items:", dst, "original:", s)

	// package slices (Go 1.21+)
	nums := []int{5, 2, 8, 1, 9}
	sorted := slices.Clone(nums)
	slices.Sort(sorted)
	fmt.Println("sorted:", sorted, "original:", nums)
	fmt.Println("contains 8:", slices.Contains(nums, 8), "index of 8:", slices.Index(nums, 8))
	fmt.Println("max:", slices.Max(nums), "min:", slices.Min(nums))
	idx, found := slices.BinarySearch(sorted, 8)
	fmt.Println("binary search 8:", idx, found)
	nums = slices.DeleteFunc(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println("odd only:", nums)
	fmt.Println("equal:", slices.Equal([]int{1, 2}, []int{1, 2}))

	// nil slice กับ empty slice: len เป็น 0 ทั้งคู่ และ append ได้เหมือนกัน
	var nilSlice []string
	empty := []string{}
	fmt.Println(nilSlice == nil, empty == nil, len(nilSlice), len(empty))
	nilSlice = append(nilSlice, "ok")
	fmt.Println(nilSlice)
}
