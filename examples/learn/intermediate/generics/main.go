// ตัวอย่าง generics: type parameter, constraint, generic type
package main

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
)

// Map แปลง slice ของ T เป็น slice ของ U
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// MaxOf ใช้ cmp.Ordered: type ที่เปรียบเทียบ < > ได้ (ตัวเลข, string)
func MaxOf[T cmp.Ordered](first T, rest ...T) T {
	m := first
	for _, v := range rest {
		m = max(m, v)
	}
	return m
}

// IndexOf ใช้ comparable: type ที่ใช้ == ได้
func IndexOf[T comparable](s []T, target T) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}

// Number เป็น constraint ที่เขียนเอง; ~ หมายถึงรวม type ที่มี underlying type นั้นด้วย
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// Baht มี underlying type เป็น int → ใช้กับ Sum ได้เพราะมี ~int
type Baht int

// Stack เป็น generic type
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

var ErrEmpty = errors.New("stack is empty")

func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if len(s.items) == 0 {
		return zero, ErrEmpty
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, nil
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Pair มี type parameter สองตัว
type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func main() {
	nums := []int{3, 1, 4, 1, 5}
	strs := Map(nums, strconv.Itoa) // Go อนุมาน T=int, U=string ให้เอง
	fmt.Printf("%q\n", strs)

	fmt.Println(MaxOf(3, 9, 2))
	fmt.Println(MaxOf("apple", "mango", "banana"))

	fmt.Println(IndexOf([]string{"a", "b", "c"}, "c"))

	prices := []Baht{120, 80, 45}
	fmt.Println("total:", Sum(prices), "baht")
	fmt.Println(Sum([]float64{1.5, 2.25}))

	var st Stack[string]
	st.Push("first")
	st.Push("second")
	for st.Len() > 0 {
		v, _ := st.Pop()
		fmt.Println("pop:", v)
	}
	if _, err := st.Pop(); err != nil {
		fmt.Println("error:", err)
	}

	p := Pair[string, int]{Key: "age", Val: 30}
	fmt.Printf("%+v\n", p)
}
