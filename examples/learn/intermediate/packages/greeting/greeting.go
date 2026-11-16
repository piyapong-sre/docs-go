// Package greeting สร้างข้อความทักทาย
package greeting

import (
	"strings"

	"learn/intermediate/packages/internal/textutil"
)

// defaultName เป็น unexported: ใช้ได้เฉพาะใน package greeting
const defaultName = "world"

// Hello เป็น exported function: ขึ้นต้นด้วยตัวพิมพ์ใหญ่
func Hello(name string) string {
	if strings.TrimSpace(name) == "" {
		name = defaultName
	}
	return "Hello, " + textutil.Title(name) + "!"
}

// Greeter เก็บคำทักทายไว้ใช้ซ้ำ
type Greeter struct {
	word string // unexported field
}

// New คืน Greeter ใหม่ (constructor function ตามธรรมเนียม Go)
func New(word string) *Greeter {
	return &Greeter{word: word}
}

// Greet ทักทายคนที่ระบุชื่อ
func (g *Greeter) Greet(name string) string {
	return g.word + " คุณ" + name
}
