// แสดงข้อมูลของ Go toolchain และเครื่องที่ใช้รันโปรแกรม
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("Go version:", runtime.Version()) // เวอร์ชันของ Go ที่ใช้ build
	fmt.Println("OS:", runtime.GOOS)              // ระบบปฏิบัติการ เช่น linux, darwin, windows
	fmt.Println("Arch:", runtime.GOARCH)          // สถาปัตยกรรม CPU เช่น amd64, arm64
	fmt.Println("CPUs:", runtime.NumCPU())        // จำนวน logical CPU
}
