// ตัวอย่าง module ที่มีหลาย package
package main

import (
	"fmt"

	"learn/intermediate/packages/greeting"
	"learn/intermediate/packages/internal/textutil"
)

func main() {
	fmt.Println(greeting.Hello("Gopher"))
	fmt.Println(greeting.Hello(""))

	g := greeting.New("สวัสดี")
	fmt.Println(g.Greet("สมชาย"))

	fmt.Println(textutil.Reverse("golang"))

	// greeting.defaultName ใช้ไม่ได้ เพราะขึ้นต้นด้วยตัวพิมพ์เล็ก (unexported)
	// fmt.Println(greeting.defaultName) // compile error
}
