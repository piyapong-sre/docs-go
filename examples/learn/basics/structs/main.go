// struct: รวมข้อมูลหลายชนิดไว้เป็นก้อนเดียว
package main

import (
	"encoding/json"
	"fmt"
)

// Address เป็น struct ธรรมดา
type Address struct {
	City    string
	Country string
}

// User มีทั้ง field ปกติ, struct tag และ embedded struct
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"` // ไม่ใส่ใน JSON ถ้าเป็นค่าว่าง
	Address         // embedding: field ของ Address ถูก "promote" ขึ้นมา
	password string // ตัวพิมพ์เล็ก = unexported, package อื่นมองไม่เห็น
}

// Point เป็น struct ที่เทียบด้วย == ได้ (ทุก field เทียบได้)
type Point struct {
	X, Y int
}

func main() {
	// struct literal แบบระบุชื่อ field (แนะนำ)
	u := User{
		ID:      1,
		Name:    "สมชาย",
		Address: Address{City: "เชียงใหม่", Country: "TH"},
	}
	fmt.Println(u.Name, u.City) // promoted field: เขียน u.City ได้เลย
	fmt.Println(u.Address.City) // หรือเขียนเต็มก็ได้
	fmt.Printf("%+v\n", u)      // %+v แสดงชื่อ field ด้วย

	// zero value ของ struct: ทุก field เป็น zero value
	var empty User
	fmt.Printf("%q %d %q\n", empty.Name, empty.ID, empty.City)

	// struct เป็น value type: assign = copy ทั้งก้อน
	u2 := u
	u2.Name = "สมหญิง"
	fmt.Println(u.Name, u2.Name)

	// anonymous struct: ใช้ครั้งเดียว ไม่ต้องตั้งชื่อ type
	config := struct {
		Host string
		Port int
	}{"localhost", 8080}
	fmt.Println(config.Host, config.Port)

	// slice ของ anonymous struct นิยมใช้ใน table-driven test
	cases := []struct {
		in   int
		want int
	}{{1, 2}, {5, 10}}
	for _, c := range cases {
		fmt.Println(c.in, "->", c.want)
	}

	// comparison
	p1, p2 := Point{1, 2}, Point{1, 2}
	fmt.Println("p1 == p2:", p1 == p2)

	// struct ใช้เป็น key ของ map ได้ (เพราะ comparable)
	visited := map[Point]bool{p1: true}
	fmt.Println(visited[Point{1, 2}], visited[Point{0, 0}])

	// struct tag: encoding/json อ่าน tag เพื่อตั้งชื่อ key
	// field password เป็น unexported จึงไม่ถูก encode
	u.password = "secret"
	data, _ := json.Marshal(u)
	fmt.Println(string(data))
}
