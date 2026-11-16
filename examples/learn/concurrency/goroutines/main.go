// ตัวอย่าง goroutine และ sync.WaitGroup
package main

import (
	"fmt"
	"sync"
	"time"
)

func download(name string, d time.Duration) {
	time.Sleep(d) // จำลองงานที่ใช้เวลา
	fmt.Printf("downloaded %s (%v)\n", name, d)
}

func main() {
	start := time.Now()

	// แบบคลาสสิก: Add ก่อน start goroutine แล้ว Done เมื่อเสร็จ
	var wg sync.WaitGroup
	files := map[string]time.Duration{
		"a.txt": 300 * time.Millisecond,
		"b.txt": 100 * time.Millisecond,
		"c.txt": 200 * time.Millisecond,
	}
	for name, d := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			download(name, d) // Go 1.22+ ตัวแปร loop เป็นของใหม่ทุกรอบ ไม่ต้อง copy
		}()
	}
	wg.Wait()
	fmt.Printf("classic: done in ~%v\n", time.Since(start).Round(100*time.Millisecond))

	// แบบใหม่ Go 1.25+: wg.Go จัดการ Add/Done ให้เอง
	start = time.Now()
	var wg2 sync.WaitGroup
	results := make([]int, 5)
	for i := range 5 {
		wg2.Go(func() {
			time.Sleep(100 * time.Millisecond)
			results[i] = i * i // แต่ละ goroutine เขียนคนละ index → ไม่ race
		})
	}
	wg2.Wait()
	fmt.Println("squares:", results)
	fmt.Printf("wg.Go: done in ~%v\n", time.Since(start).Round(100*time.Millisecond))
}
