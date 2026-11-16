// ตัวอย่างการฝัง version ตอน build ด้วย -ldflags "-X" และอ่าน build info
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// ค่าเหล่านี้ถูกแทนที่ตอน build ได้ด้วย -ldflags "-X main.version=..."
// ต้องเป็น var ระดับ package ชนิด string (ใช้กับ const ไม่ได้)
var (
	version = "dev"
	commit  = "none"
)

func main() {
	fmt.Println("version:", version)
	fmt.Println("commit: ", commit)
	fmt.Println("mode:   ", mode) // มาจากไฟล์ที่เลือกด้วย build tag
	fmt.Printf("target:  %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Go ฝังข้อมูล module และ build settings ไว้ใน binary ให้อัตโนมัติ
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("go:     ", info.GoVersion)
		fmt.Println("module: ", info.Main.Path)
		for _, s := range info.Settings {
			switch s.Key {
			case "CGO_ENABLED", "-trimpath", "-tags":
				fmt.Printf("  %s=%s\n", s.Key, s.Value)
			}
		}
	}
}
