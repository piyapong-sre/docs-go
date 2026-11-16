// ตัวอย่าง interface: implement แบบ implicit, type assertion, type switch
package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Shape เป็น interface เล็กๆ มี method เดียวที่จำเป็น
type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }

// ไม่ต้องเขียน "implements Shape" — มี method ครบก็ถือว่า implement แล้ว
func (r Rect) Area() float64   { return r.W * r.H }
func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

// String ทำให้ Rect เป็น fmt.Stringer ด้วย
func (r Rect) String() string { return fmt.Sprintf("Rect(%gx%g)", r.W, r.H) }

// TotalArea รับ interface → ใช้ได้กับทุก type ที่มี Area()
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// describe ใช้ type switch แยกตาม type จริงที่อยู่ข้างใน
func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int %d", x)
	case string:
		return fmt.Sprintf("string %q (len %d)", x, len(x))
	case fmt.Stringer:
		return "Stringer " + x.String()
	case error:
		return "error " + x.Error()
	default:
		return fmt.Sprintf("other %T", x)
	}
}

// upperWriter ห่อ io.Writer อีกตัว แล้วแปลงเป็นตัวพิมพ์ใหญ่ก่อนเขียน
type upperWriter struct{ w io.Writer }

func (u upperWriter) Write(p []byte) (int, error) {
	return u.w.Write([]byte(strings.ToUpper(string(p))))
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	fmt.Printf("total area: %.2f\n", TotalArea(shapes))

	// type assertion แบบ comma-ok ปลอดภัย ไม่ panic
	var s Shape = Rect{2, 5}
	if r, ok := s.(Rect); ok {
		fmt.Println("it's a rect, width =", r.W)
	}
	if _, ok := s.(Circle); !ok {
		fmt.Println("not a circle")
	}

	for _, v := range []any{42, "สวัสดี", Rect{1, 2}, errors.New("boom"), 3.14, nil} {
		fmt.Println(describe(v))
	}

	// io.Copy รับ io.Writer และ io.Reader — ต่อกันได้ทุก type ที่ implement
	var w io.Writer = upperWriter{os.Stdout}
	io.Copy(w, strings.NewReader("hello, interfaces\n"))
}
