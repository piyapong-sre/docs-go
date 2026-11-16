// ตัวแปร ชนิดข้อมูล zero value การแปลง type และค่าคงที่
package main

import "fmt"

// ค่าคงที่ระดับ package
const AppName = "go-th"

// Weekday เป็น type ใหม่ที่มี underlying type เป็น int
type Weekday int

// iota เริ่มที่ 0 และเพิ่มทีละ 1 ในแต่ละบรรทัดของ const block
const (
	Sunday  Weekday = iota // 0
	Monday                 // 1
	Tuesday                // 2
)

// ใช้ iota ร่วมกับนิพจน์ได้
const (
	_  = iota             // ทิ้งค่า 0
	KB = 1 << (10 * iota) // 1 << 10
	MB                    // 1 << 20
	GB                    // 1 << 30
)

func main() {
	// 1) ประกาศด้วย var แบบระบุ type และแบบให้ compiler อนุมาน
	var name string = "Gopher"
	var age = 15
	// 2) short variable declaration (ใช้ได้เฉพาะใน function)
	city := "Bangkok"
	fmt.Println(name, age, city)

	// 3) zero value: ตัวแปรที่ไม่ได้กำหนดค่าจะมีค่าเริ่มต้นเสมอ
	var i int
	var f float64
	var ok bool
	var s string
	var p *int
	fmt.Printf("int=%d float64=%g bool=%t string=%q pointer=%v\n", i, f, ok, s, p)

	// 4) ประกาศหลายตัวพร้อมกัน
	x, y := 10, 3
	fmt.Println(x/y, x%y)

	// 5) type conversion ต้องเขียนเองเสมอ
	avg := float64(x) / float64(y)
	fmt.Printf("%T %v\n", avg, avg)

	price := 99.95
	fmt.Println(int(price)) // ตัดทศนิยมทิ้ง ไม่ปัดเศษ

	big := 300
	fmt.Println(uint8(big)) // ล้น: 300 - 256 = 44

	// 6) %T ใช้ดู type ของค่า
	r := 'ก'
	var b byte = 'A'
	fmt.Printf("%T %T %T %T\n", r, b, 1.5, 2+3i)

	// 7) ค่าคงที่
	fmt.Println(AppName, Sunday, Monday, Tuesday)
	fmt.Println(KB, MB, GB)
}
