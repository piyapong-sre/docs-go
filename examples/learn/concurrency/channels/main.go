// ตัวอย่าง channel: unbuffered, buffered, close, range, directional
package main

import (
	"fmt"
	"time"
)

// producer ส่งได้อย่างเดียว (chan<- int)
func producer(n int, out chan<- int) {
	for i := range n {
		out <- i
	}
	close(out) // ผู้ส่งเป็นคนปิด channel เสมอ
}

// square รับจาก in (<-chan) แล้วส่งผลไป out
func square(in <-chan int, out chan<- int) {
	for v := range in { // วนจนกว่า in จะถูก close
		out <- v * v
	}
	close(out)
}

func main() {
	// 1) unbuffered: ผู้ส่งจะรอจนกว่าจะมีผู้รับ (synchronize กัน)
	done := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond)
		done <- "work finished"
	}()
	fmt.Println(<-done) // block จนได้ค่า

	// 2) buffered: ส่งได้ไม่ block จนกว่า buffer จะเต็ม
	queue := make(chan string, 3)
	queue <- "a"
	queue <- "b"
	fmt.Println("len:", len(queue), "cap:", cap(queue))
	fmt.Println(<-queue, <-queue)

	// 3) อ่านจาก channel ที่ปิดแล้ว ได้ zero value และ ok=false
	ch := make(chan int, 1)
	ch <- 42
	close(ch)
	v, ok := <-ch
	fmt.Println(v, ok) // 42 true (ยังมีค่าค้างใน buffer)
	v, ok = <-ch
	fmt.Println(v, ok) // 0 false (ปิดแล้วและว่าง)

	// 4) ต่อ pipeline ด้วย directional channel
	nums := make(chan int)
	squares := make(chan int)
	go producer(5, nums)
	go square(nums, squares)
	for s := range squares {
		fmt.Print(s, " ")
	}
	fmt.Println()
}
