// ตัวอย่าง defer, panic และ recover
package main

import (
	"errors"
	"fmt"
	"os"
)

func deferOrder() {
	for i := range 3 {
		defer fmt.Println("defer", i) // ทำงานแบบ LIFO: 2, 1, 0
	}
	fmt.Println("function body done")
}

func argsEvaluatedEarly() {
	x := 1
	defer fmt.Println("deferred x =", x) // x ถูกประเมินตอนนี้เลย (= 1)
	x = 2
	fmt.Println("current x =", x)
}

// double ใช้ named result + defer แก้ค่าที่จะ return ได้
func double(n int) (result int) {
	defer func() { result *= 2 }()
	return n + 1 // result = n+1 แล้ว defer คูณ 2
}

func writeFile(name string) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	// ปิดไฟล์เสมอ และไม่ทิ้ง error ตอน Close
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	_, err = f.WriteString("hello\n")
	return err
}

// safeDivide แปลง panic เป็น error ด้วย recover
func safeDivide(a, b int) (q int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil // ถ้า b == 0 จะ panic: integer divide by zero
}

func mustPositive(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("mustPositive: got %d", n))
	}
	return n
}

func main() {
	deferOrder()
	argsEvaluatedEarly()
	fmt.Println("double(4) =", double(4))

	name := os.TempDir() + "/defer-demo.txt"
	if err := writeFile(name); err != nil {
		fmt.Println("write error:", err)
	} else {
		fmt.Println("write ok")
	}
	os.Remove(name)

	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(1, 0))

	func() {
		defer func() {
			fmt.Println("recovered from:", recover())
		}()
		mustPositive(-1)
	}()

	fmt.Println("program continues normally")
}
