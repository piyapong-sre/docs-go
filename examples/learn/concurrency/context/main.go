// ตัวอย่าง context: cancel, timeout, value, cause
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// slowQuery จำลองงานที่ใช้เวลา และเคารพ ctx (ยกเลิกได้)
func slowQuery(ctx context.Context, d time.Duration) (string, error) {
	select {
	case <-time.After(d):
		return "result", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// worker วนทำงานจนกว่า ctx จะถูกยกเลิก
func worker(ctx context.Context, id int, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d stop: %v\n", id, ctx.Err())
			return
		case <-time.After(50 * time.Millisecond):
			// ทำงานไปเรื่อยๆ
		}
	}
}

// key type แบบ unexported กันชนกับ package อื่น
type ctxKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

var ErrShutdown = errors.New("server shutting down")

func main() {
	root := context.Background()

	// 1) WithTimeout: งานเกินเวลาจะถูกยกเลิก
	ctx, cancel := context.WithTimeout(root, 100*time.Millisecond)
	defer cancel()
	_, err := slowQuery(ctx, 500*time.Millisecond)
	fmt.Println("timeout query:", err)
	fmt.Println("is DeadlineExceeded?", errors.Is(err, context.DeadlineExceeded))

	// 2) WithCancel: ยกเลิกแม่ ลูกทุกตัวหยุดตาม
	parent, cancelParent := context.WithCancel(root)
	done := make(chan struct{})
	for i := range 2 {
		child, cancelChild := context.WithCancel(parent)
		defer cancelChild()
		go worker(child, i+1, done)
	}
	time.Sleep(120 * time.Millisecond)
	cancelParent()
	<-done
	<-done

	// 3) WithDeadline: กำหนดเวลาสิ้นสุดเป็นเวลาจริง
	dl := time.Now().Add(50 * time.Millisecond)
	ctx2, cancel2 := context.WithDeadline(root, dl)
	defer cancel2()
	deadline, ok := ctx2.Deadline()
	fmt.Println("has deadline:", ok, "in future:", deadline.After(time.Now()))
	<-ctx2.Done()
	fmt.Println("deadline ctx:", ctx2.Err())

	// 4) WithValue: ใช้เฉพาะข้อมูลระดับ request เช่น request ID
	ctx3 := withRequestID(root, "req-123")
	fmt.Println("request id:", requestID(ctx3))

	// 5) WithCancelCause: บอกเหตุผลที่ยกเลิก
	ctx4, cancel4 := context.WithCancelCause(root)
	cancel4(ErrShutdown)
	fmt.Println("err:", ctx4.Err())
	fmt.Println("cause:", context.Cause(ctx4))
}
