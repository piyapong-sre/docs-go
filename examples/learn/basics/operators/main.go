// ตัวดำเนินการ (operators) ใน Go
package main

import "fmt"

func main() {
	a, b := 17, 5

	// ตัวดำเนินการทางคณิตศาสตร์
	fmt.Println("a + b =", a+b)
	fmt.Println("a - b =", a-b)
	fmt.Println("a * b =", a*b)
	fmt.Println("a / b =", a/b) // int หาร int ได้ int (ปัดทิ้งไปทาง 0)
	fmt.Println("a % b =", a%b)
	fmt.Println("-a / b =", -a/b, " -a % b =", -a%b)
	fmt.Println("17.0 / 5 =", 17.0/5)

	// ตัวดำเนินการเปรียบเทียบ ให้ผลเป็น bool
	fmt.Println(a == b, a != b, a > b, a <= b)

	// ตัวดำเนินการทางตรรกะ (short-circuit)
	age, hasTicket := 20, false
	fmt.Println(age >= 18 && hasTicket, age >= 18 || hasTicket, !hasTicket)

	// ตัวดำเนินการระดับ bit
	x, y := 0b1100, 0b1010
	fmt.Printf("x & y  = %04b\n", x&y)  // AND
	fmt.Printf("x | y  = %04b\n", x|y)  // OR
	fmt.Printf("x ^ y  = %04b\n", x^y)  // XOR
	fmt.Printf("x &^ y = %04b\n", x&^y) // AND NOT (bit clear)
	fmt.Println("1 << 4 =", 1<<4, " 256 >> 2 =", 256>>2)

	// assignment operators และ ++ / --
	n := 10
	n += 5
	n *= 2
	n--
	fmt.Println("n =", n)

	// ต่อ string ด้วย +
	greeting := "สวัสดี" + " " + "Go"
	fmt.Println(greeting)

	// min / max เป็น built-in (Go 1.21+)
	fmt.Println(min(3, 1, 2), max(2.5, 7.1))

	// integer overflow: ค่าวนรอบ ไม่มี error
	var small int8 = 127
	small++
	fmt.Println("int8 127 + 1 =", small)
}
