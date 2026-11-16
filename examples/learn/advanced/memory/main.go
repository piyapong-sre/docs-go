// ตัวอย่างการลด allocation: escape analysis, preallocate, strings.Builder, sync.Pool
package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type Point struct{ X, Y int }

// newPointValue คืนค่าแบบ value: ค่าอยู่บน stack ได้ ไม่ต้อง allocate
func newPointValue(x, y int) Point {
	return Point{x, y}
}

// newPointPtr คืน pointer ไปยังตัวแปร local -> ตัวแปร "escape" ไป heap
//
//go:noinline
func newPointPtr(x, y int) *Point {
	p := Point{x, y}
	return &p
}

// appendNoPrealloc ปล่อยให้ slice โตเอง -> allocate ใหม่หลายรอบ
func appendNoPrealloc(n int) []int {
	var s []int
	for i := range n {
		s = append(s, i)
	}
	return s
}

// appendPrealloc รู้ขนาดล่วงหน้า -> allocate ครั้งเดียว
func appendPrealloc(n int) []int {
	s := make([]int, 0, n)
	for i := range n {
		s = append(s, i)
	}
	return s
}

// concatPlus ต่อ string ด้วย + ทุกรอบสร้าง string ใหม่
func concatPlus(words []string) string {
	var s string
	for _, w := range words {
		s += w + ","
	}
	return s
}

// concatBuilder ใช้ strings.Builder + Grow
func concatBuilder(words []string) string {
	var b strings.Builder
	b.Grow(len(words) * 8)
	for _, w := range words {
		b.WriteString(w)
		b.WriteByte(',')
	}
	return b.String()
}

// bufPool เก็บ *bytes.Buffer ไว้ใช้ซ้ำ ลดภาระ GC ใน hot path
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func render(name string) int {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset() // สำคัญ: ล้างข้อมูลเก่าก่อนใช้
	defer bufPool.Put(buf)

	buf.WriteString("hello, ")
	buf.WriteString(name)
	return buf.Len()
}

// ตัวแปร sink ระดับ package กันไม่ให้ compiler ตัดโค้ดทิ้ง
// และทำให้ผลลัพธ์ "หนี" ออกจากฟังก์ชันเหมือนการใช้งานจริง
var (
	sinkPoint *Point
	sinkInts  []int
	sinkStr   string
)

func main() {
	words := strings.Fields("go is a simple fast and fun language")

	// testing.AllocsPerRun นับจำนวน allocation เฉลี่ยต่อการเรียก 1 ครั้ง
	show := func(name string, f func()) {
		fmt.Printf("%-18s %4.0f allocs/op\n", name, testing.AllocsPerRun(100, f))
	}
	show("newPointValue", func() { _ = newPointValue(1, 2) })
	show("newPointPtr", func() { sinkPoint = newPointPtr(1, 2) })
	show("appendNoPrealloc", func() { sinkInts = appendNoPrealloc(1000) })
	show("appendPrealloc", func() { sinkInts = appendPrealloc(1000) })
	show("concatPlus", func() { sinkStr = concatPlus(words) })
	show("concatBuilder", func() { sinkStr = concatBuilder(words) })
	show("render (sync.Pool)", func() { _ = render("gopher") })
}
