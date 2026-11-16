// ตัวอย่าง "ชิมลาง" ภาษา Go: type, function, slice, error และ goroutine
package main

import (
	"fmt"
	"strings"
	"sync"
)

// Book เก็บข้อมูลหนังสือหนึ่งเล่ม
type Book struct {
	Title string
	Pages int
}

func main() {
	books := []Book{
		{Title: "Go เบื้องต้น", Pages: 120},
		{Title: "Concurrency in Go", Pages: 240},
		{Title: "Web API ด้วย Go", Pages: 310},
	}

	// สร้าง slice สำหรับเก็บผลลัพธ์ แต่ละ goroutine เขียนช่องของตัวเอง
	summaries := make([]string, len(books))

	var wg sync.WaitGroup
	for i, b := range books {
		// wg.Go (Go 1.25+) สั่งให้ function ทำงานใน goroutine ใหม่
		wg.Go(func() {
			summaries[i] = fmt.Sprintf("- %s (%d หน้า)", b.Title, b.Pages)
		})
	}
	wg.Wait() // รอทุก goroutine ทำงานเสร็จ

	fmt.Println(strings.Join(summaries, "\n"))
}
