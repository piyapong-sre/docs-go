package calc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// test พื้นฐาน: ชื่อขึ้นต้นด้วย Test และรับ *testing.T
func TestAdd(t *testing.T) {
	got := Add(2, 3)
	if got != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", got)
	}
}

// table-driven test + subtest ด้วย t.Run
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr error
	}{
		{name: "even", a: 10, b: 2, want: 5},
		{name: "truncate", a: 7, b: 2, want: 3},
		{name: "negative", a: -9, b: 3, want: -3},
		{name: "by zero", a: 1, b: 0, wantErr: ErrDivideByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // subtest แต่ละตัวรันขนานกันได้ (Go 1.22+ ไม่ต้อง copy tt แล้ว)

			got, err := Divide(tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Divide(%d, %d) error = %v; want %v", tt.a, tt.b, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// assertEqual เป็น helper: t.Helper() ทำให้ error ชี้ไปที่บรรทัดของผู้เรียก
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestReverse(t *testing.T) {
	assertEqual(t, Reverse("abc"), "cba")
	assertEqual(t, Reverse("สวัสดี"), "ีดสัวส")
	assertEqual(t, Reverse(""), "")
}

// อ่านไฟล์ทดสอบจากโฟลเดอร์ testdata/ (go tool จะไม่ compile โฟลเดอร์นี้)
func TestWordCountFromFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "sample.txt"))
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, WordCount(string(data)), 11)
}

// benchmark แบบใหม่ด้วย b.Loop() (Go 1.24+)
func BenchmarkReverse(b *testing.B) {
	s := strings.Repeat("สวัสดี golang ", 20)
	for b.Loop() {
		Reverse(s)
	}
}

// fuzz test: Go จะสุ่ม input มาทดสอบ property ที่ต้องเป็นจริงเสมอ
func FuzzReverse(f *testing.F) {
	f.Add("hello")
	f.Add("สวัสดี")
	f.Fuzz(func(t *testing.T, s string) {
		if !IsValidUTF8(s) {
			t.Skip("ข้าม input ที่ไม่ใช่ UTF-8")
		}
		if got := Reverse(Reverse(s)); got != s {
			t.Errorf("Reverse twice = %q; want %q", got, s)
		}
	})
}
