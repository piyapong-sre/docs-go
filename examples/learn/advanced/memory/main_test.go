package main

import (
	"strings"
	"testing"
)

var words = strings.Fields("go is a simple fast and fun language")

// เก็บผลลัพธ์ไว้ใน sink กัน compiler optimize ทิ้ง

func BenchmarkAppendNoPrealloc(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkInts = appendNoPrealloc(1000)
	}
}

func BenchmarkAppendPrealloc(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkInts = appendPrealloc(1000)
	}
}

func BenchmarkConcatPlus(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = concatPlus(words)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkStr = concatBuilder(words)
	}
}

func TestConcatSame(t *testing.T) {
	if concatPlus(words) != concatBuilder(words) {
		t.Fatal("ผลลัพธ์ต้องเท่ากัน")
	}
}
