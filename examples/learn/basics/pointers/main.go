// pointer: ตัวแปรที่เก็บ "ที่อยู่" ของค่าอื่น
package main

import "fmt"

type Counter struct {
	Name  string
	Value int
}

// รับ value: ได้สำเนา แก้แล้วไม่กระทบต้นฉบับ
func resetByValue(c Counter) {
	c.Value = 0
}

// รับ pointer: แก้ค่าต้นฉบับได้
func resetByPointer(c *Counter) {
	c.Value = 0 // Go แปลง c.Value เป็น (*c).Value ให้อัตโนมัติ
}

// คืน pointer ของตัวแปร local ได้อย่างปลอดภัย
// (compiler ย้ายค่าไปไว้บน heap ให้เอง: escape analysis)
func newCounter(name string) *Counter {
	c := Counter{Name: name}
	return &c
}

func double(n *int) {
	*n *= 2
}

func main() {
	x := 10
	p := &x // & เอาที่อยู่ของ x

	fmt.Printf("type of p: %T\n", p)
	fmt.Println("*p =", *p) // * อ่านค่าที่ pointer ชี้ (dereference)

	*p = 20 // เขียนค่าผ่าน pointer
	fmt.Println("x =", x)

	double(&x)
	fmt.Println("after double:", x)

	// value vs pointer
	c := Counter{Name: "visits", Value: 42}
	resetByValue(c)
	fmt.Println("after resetByValue:", c.Value)
	resetByPointer(&c)
	fmt.Println("after resetByPointer:", c.Value)

	// new(T): จองพื้นที่ให้ค่า zero value ของ T แล้วคืน *T
	n := new(int)
	fmt.Println("new(int):", *n)
	*n = 5
	fmt.Println("*n =", *n)

	// &T{} นิยมกว่า new สำหรับ struct
	cp := &Counter{Name: "clicks"}
	cp.Value++
	fmt.Printf("%+v\n", *cp)

	nc := newCounter("likes")
	nc.Value += 3
	fmt.Println(nc.Name, nc.Value)

	// สอง pointer ชี้ที่เดียวกัน
	a := &Counter{Name: "shared"}
	b := a
	b.Value = 99
	fmt.Println("a.Value =", a.Value, " same:", a == b)

	// zero value ของ pointer คือ nil: ต้องเช็คก่อนใช้
	var np *Counter
	if np == nil {
		fmt.Println("np is nil")
	}
	// fmt.Println(np.Value) // panic: invalid memory address or nil pointer dereference
}
