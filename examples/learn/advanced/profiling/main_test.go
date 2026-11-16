package main

import (
	"maps"
	"testing"
)

func TestWordFreqSame(t *testing.T) {
	text := sampleText()
	if !maps.Equal(wordFreqSlow(text), wordFreqFast(text)) {
		t.Fatal("ผลลัพธ์ของสองเวอร์ชันต้องเท่ากัน")
	}
}

func BenchmarkWordFreqSlow(b *testing.B) {
	text := sampleText()
	b.ReportAllocs()
	for b.Loop() {
		wordFreqSlow(text)
	}
}

func BenchmarkWordFreqFast(b *testing.B) {
	text := sampleText()
	b.ReportAllocs()
	for b.Loop() {
		wordFreqFast(text)
	}
}
