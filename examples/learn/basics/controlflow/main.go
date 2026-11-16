// control flow: if, for, switch, break/continue และ label
package main

import (
	"fmt"
	"strconv"
)

func main() {
	// if พร้อม init statement: ตัวแปร n, err มี scope แค่ใน if/else นี้
	if n, err := strconv.Atoi("42"); err != nil {
		fmt.Println("แปลงไม่ได้:", err)
	} else {
		fmt.Println("ได้ตัวเลข", n)
	}

	// for แบบ 3 ส่วน
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// for แบบ while
	total := 1
	for total < 100 {
		total *= 3
	}
	fmt.Println("total =", total)

	// range over int (Go 1.22+): วน 0, 1, 2
	for i := range 3 {
		fmt.Print("รอบที่ ", i, "; ")
	}
	fmt.Println()

	// range over slice ได้ทั้ง index และค่า
	fruits := []string{"มะม่วง", "ทุเรียน", "มังคุด"}
	for i, f := range fruits {
		fmt.Println(i, f)
	}

	// for ไม่มีเงื่อนไข (วนไม่รู้จบ) ต้องมี break
	count := 0
	for {
		count++
		if count%2 == 0 {
			continue // ข้ามเลขคู่
		}
		if count > 7 {
			break
		}
		fmt.Print(count, " ")
	}
	fmt.Println()

	// switch: ไม่ต้องเขียน break, case เดียวรับหลายค่าได้
	day := "sat"
	switch day {
	case "sat", "sun":
		fmt.Println("วันหยุด")
	case "mon":
		fmt.Println("วันจันทร์อีกแล้ว")
	default:
		fmt.Println("วันทำงาน")
	}

	// switch แบบไม่มีนิพจน์ = if-else หลายชั้นที่อ่านง่ายกว่า
	score := 72
	switch {
	case score >= 80:
		fmt.Println("เกรด A")
	case score >= 70:
		fmt.Println("เกรด B")
	default:
		fmt.Println("เกรด C")
	}

	// fallthrough: บังคับให้ทำ case ถัดไปต่อ (ใช้น้อยมาก)
	switch n := 1; n {
	case 1:
		fmt.Println("one")
		fallthrough
	case 2:
		fmt.Println("two")
	case 3:
		fmt.Println("three")
	}

	// label: break ออกจาก loop ชั้นนอกโดยตรง
	matrix := [][]int{{1, 2, 3}, {4, -5, 6}, {7, 8, 9}}
outer:
	for r, row := range matrix {
		for c, v := range row {
			if v < 0 {
				fmt.Printf("เจอค่าติดลบที่ [%d][%d]\n", r, c)
				break outer
			}
		}
	}
}
