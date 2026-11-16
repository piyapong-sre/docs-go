// ตัวอย่างแก้ data race และ goroutine leak (เวอร์ชันที่ถูกต้องแล้ว)
package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ---- data race: แก้ด้วย Mutex ----

type Stats struct {
	mu    sync.Mutex
	count int
}

func (s *Stats) Inc() {
	s.mu.Lock()
	s.count++
	s.mu.Unlock()
}

// ---- data race: แก้ด้วย atomic ----

func countAtomic(n int) int64 {
	var total atomic.Int64
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() { total.Add(1) })
	}
	wg.Wait()
	return total.Load()
}

// ---- goroutine leak: แก้ด้วย buffered channel ----

// firstResult ถามหลายแหล่งแต่ใช้คำตอบแรกเท่านั้น
// ถ้าใช้ unbuffered channel goroutine ที่ตอบช้าจะค้างตลอดไป (leak)
func firstResult(sources []string) string {
	ch := make(chan string, len(sources)) // buffer พอสำหรับทุกคน → ไม่มีใครค้าง
	for _, s := range sources {
		go func() {
			time.Sleep(10 * time.Millisecond)
			ch <- s
		}()
	}
	return <-ch
}

// ---- goroutine leak: แก้ด้วย context ----

func ticker(ctx context.Context, out chan<- int) {
	for i := 0; ; i++ {
		select {
		case out <- i:
		case <-ctx.Done(): // มีทางออกเสมอ
			return
		}
	}
}

func main() {
	var s Stats
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(s.Inc)
	}
	wg.Wait()
	fmt.Println("mutex count:", s.count)
	fmt.Println("atomic count:", countAtomic(1000))

	before := runtime.NumGoroutine()
	for range 10 {
		firstResult([]string{"a", "b", "c"})
	}

	ctx, cancel := context.WithCancel(context.Background())
	nums := make(chan int)
	go ticker(ctx, nums)
	for range 3 {
		<-nums
	}
	cancel() // ถ้าลืมบรรทัดนี้ goroutine ticker จะค้างตลอดไป

	time.Sleep(50 * time.Millisecond) // รอให้ goroutine จบ
	fmt.Println("leaked goroutines:", runtime.NumGoroutine()-before)
}
