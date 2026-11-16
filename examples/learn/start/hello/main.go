// ทุกไฟล์ Go ต้องเริ่มด้วยการบอกว่าอยู่ใน package ไหน
// package main คือ package พิเศษที่ compile เป็นโปรแกรมที่รันได้
package main

// import package จาก standard library ที่ต้องใช้
import (
	"fmt"
	"os"
)

// func main คือจุดเริ่มต้นของโปรแกรม (entry point)
func main() {
	fmt.Println("สวัสดี Go!")

	name := "Gopher"
	if len(os.Args) > 1 {
		name = os.Args[1] // argument แรกที่ส่งมาจาก command line
	}
	fmt.Printf("ยินดีต้อนรับ %s\n", name)
}
