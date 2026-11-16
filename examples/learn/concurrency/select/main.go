// ตัวอย่าง select: รอหลาย channel, default, timeout, ticker
package main

import (
	"fmt"
	"time"
)

func fetch(name string, d time.Duration) <-chan string {
	ch := make(chan string, 1) // buffer 1 → goroutine ไม่ค้างแม้ไม่มีใครรับ
	go func() {
		time.Sleep(d)
		ch <- name
	}()
	return ch
}

func main() {
	// 1) รอว่าใครมาก่อน
	fast := fetch("mirror-asia", 50*time.Millisecond)
	slow := fetch("mirror-eu", 200*time.Millisecond)
	select {
	case r := <-fast:
		fmt.Println("first:", r)
	case r := <-slow:
		fmt.Println("first:", r)
	}

	// 2) timeout ด้วย time.After
	select {
	case r := <-fetch("slow-api", 300*time.Millisecond):
		fmt.Println("got", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timeout!")
	}

	// 3) default: ไม่ block ถ้ายังไม่มีอะไรพร้อม
	msgs := make(chan string, 1)
	select {
	case m := <-msgs:
		fmt.Println("message:", m)
	default:
		fmt.Println("no message, moving on")
	}

	// 4) Ticker: ทำงานเป็นจังหวะ จนกว่าจะได้สัญญาณหยุด
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	stop := time.After(350 * time.Millisecond)
	count := 0
	for {
		select {
		case <-ticker.C:
			count++
			fmt.Println("tick", count)
		case <-stop:
			fmt.Println("stopped after", count, "ticks")
			return
		}
	}
}
