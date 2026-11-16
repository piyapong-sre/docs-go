// map: key-value
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func main() {
	// สร้างด้วย literal
	stock := map[string]int{
		"apple":  5,
		"banana": 0,
		"mango":  12,
	}

	// สร้างด้วย make
	prices := make(map[string]float64)
	prices["apple"] = 25.5
	prices["mango"] = 40

	// อ่านค่า: key ที่ไม่มีจะได้ zero value
	fmt.Println("apple:", stock["apple"], "grape:", stock["grape"])

	// comma-ok idiom: แยก "ไม่มี key" ออกจาก "มี key แต่ค่าเป็น 0"
	if qty, ok := stock["banana"]; ok {
		fmt.Println("banana มีใน map, qty =", qty)
	}
	if _, ok := stock["grape"]; !ok {
		fmt.Println("ไม่มี grape ใน map")
	}

	// เพิ่ม/แก้ไข และลบ
	stock["grape"] = 7
	stock["apple"] += 3
	delete(stock, "banana")
	delete(stock, "durian") // ลบ key ที่ไม่มีก็ไม่ error
	fmt.Println("len:", len(stock))

	// ลำดับการวน map เป็นแบบสุ่ม ถ้าต้องการลำดับคงที่ให้เรียง key ก่อน
	for _, k := range slices.Sorted(maps.Keys(stock)) {
		fmt.Printf("%-6s %d\n", k, stock[k])
	}

	// fmt.Println เรียง key ให้อัตโนมัติ (เฉพาะตอนพิมพ์)
	fmt.Println(stock)

	// ตัวอย่างใช้งานจริง: นับจำนวนคำ
	text := "go is fun and go is fast"
	counts := map[string]int{}
	for _, w := range strings.Fields(text) {
		counts[w]++ // key ใหม่เริ่มจาก 0 แล้วบวก 1
	}
	fmt.Println(counts)

	// map ที่มี value เป็น slice
	groups := map[string][]string{}
	groups["ผลไม้"] = append(groups["ผลไม้"], "มะม่วง", "ทุเรียน")
	groups["ผัก"] = append(groups["ผัก"], "คะน้า")
	fmt.Println(len(groups["ผลไม้"]), groups["ผัก"])

	// package maps (Go 1.21+)
	backup := maps.Clone(stock)
	maps.DeleteFunc(backup, func(k string, v int) bool { return v < 10 })
	fmt.Println("backup:", backup, "stock:", stock)
	fmt.Println("equal:", maps.Equal(stock, backup))

	// clear ลบทุก key (Go 1.21+)
	clear(backup)
	fmt.Println("after clear:", backup, len(backup))

	// nil map: อ่านได้ แต่เขียนไม่ได้
	var nilMap map[string]int
	fmt.Println(nilMap == nil, nilMap["x"], len(nilMap))
	// nilMap["x"] = 1 // panic: assignment to entry in nil map
}
