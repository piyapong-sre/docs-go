// ตัวอย่าง profiling: เขียน CPU และ heap profile ด้วย runtime/pprof
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

// wordFreqSlow นับคำแบบไม่ระวัง: สร้าง string ใหม่ทุกรอบด้วย ToLower + Fields
func wordFreqSlow(text string) map[string]int {
	counts := map[string]int{}
	for _, line := range strings.Split(text, "\n") {
		for _, w := range strings.Fields(strings.ToLower(line)) {
			counts[w]++
		}
	}
	return counts
}

// wordFreqFast ลดงานซ้ำ: lower ครั้งเดียว, ใช้ FieldsSeq (Go 1.24+) ไม่ต้องสร้าง slice
func wordFreqFast(text string) map[string]int {
	counts := make(map[string]int, 64)
	for w := range strings.FieldsSeq(strings.ToLower(text)) {
		counts[w]++
	}
	return counts
}

func sampleText() string {
	return strings.Repeat("Go is fun\nGo is FAST and go is simple\n", 2000)
}

func main() {
	dir := os.TempDir()
	cpuPath := filepath.Join(dir, "cpu.pprof")
	memPath := filepath.Join(dir, "mem.pprof")

	// --- CPU profile: เริ่มเก็บ -> ทำงาน -> หยุด ---
	f, err := os.Create(cpuPath)
	if err != nil {
		panic(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	text := sampleText()
	var total int
	for range 200 {
		total += len(wordFreqSlow(text))
		total += len(wordFreqFast(text))
	}
	pprof.StopCPUProfile()
	f.Close()

	// --- Heap profile: snapshot ของ memory ณ ตอนนี้ ---
	m, err := os.Create(memPath)
	if err != nil {
		panic(err)
	}
	runtime.GC() // ให้ข้อมูล heap เป็นปัจจุบัน
	if err := pprof.WriteHeapProfile(m); err != nil {
		panic(err)
	}
	m.Close()

	fmt.Println("unique words:", len(wordFreqFast(text)))
	fmt.Println("wrote", filepath.Base(cpuPath), "and", filepath.Base(memPath), "to temp dir")
	fmt.Println("ดูผลด้วย: go tool pprof -top <temp dir>/cpu.pprof")
	_ = total
}
