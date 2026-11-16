// string, byte, rune และ UTF-8 (เน้นข้อความภาษาไทย)
package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// truncateRunes ตัดข้อความให้เหลือไม่เกิน n rune (ไม่ทำให้ตัวอักษรแตก)
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// truncateThai ตัดให้เหลือ n "ตัวอักษรหลัก" และเก็บสระบน/ล่าง/วรรณยุกต์
// (combining mark, หมวด unicode.Mn) ที่ตามหลังไว้ด้วย
func truncateThai(s string, n int) string {
	count := 0
	for i, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue // mark ติดไปกับตัวอักษรก่อนหน้า ไม่นับ
		}
		if count == n {
			return s[:i] // i เป็น byte index ที่อยู่ตรงขอบ rune เสมอ
		}
		count++
	}
	return s
}

func main() {
	s := "สวัสดี"

	// len() นับเป็น byte ไม่ใช่จำนวนตัวอักษร
	fmt.Println("len:", len(s))
	fmt.Println("runes:", utf8.RuneCountInString(s))
	fmt.Println("len(\"Go\"):", len("Go"))

	// s[i] ได้ byte ไม่ใช่ตัวอักษร
	fmt.Printf("s[0] = %d (%T)\n", s[0], s[0])
	fmt.Printf("bytes: % x\n", s)

	// range over string: ได้ byte index และ rune
	for i, r := range "กข้" {
		fmt.Printf("index %d: %c U+%04X\n", i, r, r)
	}

	// แปลงเป็น []rune เพื่อเข้าถึงทีละตัวอักษร
	runes := []rune(s)
	fmt.Println("rune[1]:", string(runes[1]), "count:", len(runes))

	// ตัดข้อความ: ตัดตาม byte อาจได้ UTF-8 ที่พัง
	bad := s[:4]
	fmt.Printf("s[:4] = %q valid=%t\n", bad, utf8.ValidString(bad))
	fmt.Println("truncateRunes:", truncateRunes(s, 3))
	fmt.Println("truncateRunes:", truncateRunes("ที่นี่", 2)) // วรรณยุกต์หาย!
	fmt.Println("truncateThai: ", truncateThai("ที่นี่", 1))
	fmt.Println("truncateThai: ", truncateThai("ที่นี่ประเทศไทย", 4))

	// package strings
	text := "  Go เป็นภาษาที่เรียบง่าย  "
	t := strings.TrimSpace(text)
	fmt.Printf("%q\n", t)
	fmt.Println(strings.Contains(t, "ภาษา"), strings.HasPrefix(t, "Go"))
	fmt.Println(strings.Index(t, "เป็น")) // byte index!
	fmt.Println(strings.ToUpper("hello"), strings.Repeat("ab", 3))
	fmt.Println(strings.ReplaceAll("a-b-c", "-", "/"))
	fmt.Printf("%q\n", strings.Split("a,b,c", ","))
	fmt.Printf("%q\n", strings.Fields(" one  two\tthree "))
	fmt.Println(strings.Join([]string{"กรุงเทพ", "เชียงใหม่"}, " | "))
	before, after, found := strings.Cut("key=value", "=")
	fmt.Println(before, after, found)

	// strings.Builder: ต่อ string จำนวนมากอย่างมีประสิทธิภาพ
	var sb strings.Builder
	for i := range 3 {
		fmt.Fprintf(&sb, "[%d]", i)
	}
	sb.WriteString(" เสร็จ")
	fmt.Println(sb.String())

	// package strconv: แปลง string <-> ตัวเลข
	n, err := strconv.Atoi("123")
	fmt.Println(n+1, err)
	_, err = strconv.Atoi("12a")
	fmt.Println(err)
	f, _ := strconv.ParseFloat("3.14", 64)
	fmt.Println(f * 2)
	fmt.Println(strconv.Itoa(42)+"!", strconv.FormatInt(255, 2), strconv.Quote("ไทย\n"))
	ok, _ := strconv.ParseBool("true")
	fmt.Println(ok)

	// raw string literal ใช้ backtick: ไม่ตีความ escape
	path := `C:\new\folder`
	fmt.Println(path)
}
