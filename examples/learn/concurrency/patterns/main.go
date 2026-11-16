// ตัวอย่าง concurrency patterns: worker pool, pipeline, fan-out/fan-in,
// errgroup และ semaphore
package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ---------- 1) Worker pool ----------

type Job struct{ ID, N int }
type Result struct{ JobID, Value int }

func workerPool(jobs []Job, workers int) []Result {
	jobCh := make(chan Job)
	resCh := make(chan Result)

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for j := range jobCh {
				time.Sleep(10 * time.Millisecond) // จำลองงาน
				resCh <- Result{JobID: j.ID, Value: j.N * j.N}
			}
		})
	}

	// ส่งงานเข้าคิว แล้วปิด channel เมื่อส่งครบ
	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
	}()

	// ปิด resCh เมื่อ worker ทุกตัวจบ
	go func() {
		wg.Wait()
		close(resCh)
	}()

	var results []Result
	for r := range resCh {
		results = append(results, r)
	}
	// ผลลัพธ์มาไม่เรียงลำดับ → sort ก่อนคืน
	slices.SortFunc(results, func(a, b Result) int { return a.JobID - b.JobID })
	return results
}

// ---------- 2) Pipeline + fan-out/fan-in ----------

func generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// merge (fan-in) รวมหลาย channel เป็น channel เดียว
func merge(ctx context.Context, chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Go(func() {
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// ---------- 3) errgroup ----------

func fetchAll(ctx context.Context, urls []string) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(2) // รันพร้อมกันไม่เกิน 2 งาน

	results := make([]string, len(urls))
	for i, u := range urls {
		g.Go(func() error {
			if u == "bad://url" {
				return fmt.Errorf("fetch %s: %w", u, errors.ErrUnsupported)
			}
			select {
			case <-time.After(20 * time.Millisecond):
				results[i] = "body of " + u
				return nil
			case <-ctx.Done(): // งานอื่นพังแล้ว → หยุด
				return ctx.Err()
			}
		})
	}
	if err := g.Wait(); err != nil { // คืน error แรกที่เกิด
		return nil, err
	}
	return results, nil
}

// ---------- 4) Semaphore ด้วย buffered channel ----------

func limitedWork(tasks, limit int) int64 {
	sem := make(chan struct{}, limit)
	var mu sync.Mutex
	var running, peak int64
	var wg sync.WaitGroup
	for range tasks {
		wg.Go(func() {
			sem <- struct{}{}        // ขอ slot (block ถ้าเต็ม)
			defer func() { <-sem }() // คืน slot

			mu.Lock()
			running++
			peak = max(peak, running)
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			running--
			mu.Unlock()
		})
	}
	wg.Wait()
	return peak
}

func main() {
	jobs := []Job{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}}
	fmt.Println("worker pool:", workerPool(jobs, 3))

	ctx := context.Background()
	src := generate(ctx, 1, 2, 3, 4, 5, 6)
	// fan-out: ให้ square 3 ตัวแย่งกันอ่านจาก src
	sq1, sq2, sq3 := square(ctx, src), square(ctx, src), square(ctx, src)
	var sums []int
	for v := range merge(ctx, sq1, sq2, sq3) { // fan-in
		sums = append(sums, v)
	}
	slices.Sort(sums)
	fmt.Println("pipeline:", sums)

	res, err := fetchAll(ctx, []string{"a.com", "b.com", "c.com"})
	fmt.Println("errgroup ok:", res, err)
	_, err = fetchAll(ctx, []string{"a.com", "bad://url", "c.com"})
	fmt.Println("errgroup err:", err)

	fmt.Println("semaphore peak concurrency:", limitedWork(10, 3))
}
