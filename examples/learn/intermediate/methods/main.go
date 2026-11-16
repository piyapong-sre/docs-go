// ตัวอย่าง method: value receiver vs pointer receiver
package main

import "fmt"

// Rect เป็น struct ขนาดเล็ก ไม่ต้องแก้ค่า → ใช้ value receiver ได้
type Rect struct {
	Width, Height float64
}

// Area ใช้ value receiver: ได้ "สำเนา" ของ Rect มาคำนวณ
func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// Counter ต้องแก้ค่าภายใน → ใช้ pointer receiver
type Counter struct {
	name  string
	count int
}

// Inc ใช้ pointer receiver: แก้ค่าตัวจริงได้
func (c *Counter) Inc() {
	c.count++
}

// IncWrong ใช้ value receiver: แก้แค่สำเนา ตัวจริงไม่เปลี่ยน
func (c Counter) IncWrong() {
	c.count++
}

// String ใช้ pointer receiver ให้สอดคล้องกับ method อื่นของ Counter
func (c *Counter) String() string {
	return fmt.Sprintf("%s=%d", c.name, c.count)
}

// Celsius เป็น named type จาก float64 ก็มี method ได้
type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

func main() {
	r := Rect{Width: 3, Height: 4}
	fmt.Println("area:", r.Area())

	c := Counter{name: "visits"}
	c.Inc() // Go แปลงเป็น (&c).Inc() ให้อัตโนมัติ เพราะ c เป็นตัวแปรที่ address ได้
	c.Inc()
	c.IncWrong() // ไม่มีผลกับ c
	fmt.Println(c.String())

	p := &Counter{name: "clicks"}
	p.Inc()
	fmt.Println(p) // fmt เรียก String() ให้ เพราะ *Counter implement fmt.Stringer

	t := Celsius(36.6)
	fmt.Printf("%.1f°C = %.2f°F\n", float64(t), t.Fahrenheit())

	// method value: ผูก method กับตัวแปรไว้ เรียกทีหลังได้
	inc := c.Inc
	inc()
	inc()
	fmt.Println(c.String())
}
