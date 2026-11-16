// ตัวอย่าง sync: Mutex, RWMutex, Once, OnceValue, atomic
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// SafeCounter ใช้ Mutex ป้องกัน map จากการเขียนพร้อมกัน
type SafeCounter struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key]++
}

func (c *SafeCounter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

// Cache อ่านบ่อย เขียนน้อย → RWMutex ให้หลาย reader อ่านพร้อมกันได้
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *Cache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}

// loadConfig จะถูกเรียกครั้งเดียวไม่ว่าเรียก getConfig กี่ครั้ง
var getConfig = sync.OnceValue(func() map[string]string {
	fmt.Println("loading config (once)")
	return map[string]string{"env": "dev"}
})

func main() {
	var wg sync.WaitGroup

	// Mutex
	c := SafeCounter{m: make(map[string]int)}
	for range 1000 {
		wg.Go(func() { c.Inc("hits") })
	}
	wg.Wait()
	fmt.Println("hits:", c.Get("hits"))

	// RWMutex
	cache := Cache{data: make(map[string]string)}
	cache.Set("lang", "go")
	for range 5 {
		wg.Go(func() { cache.Get("lang") })
	}
	wg.Wait()
	v, _ := cache.Get("lang")
	fmt.Println("lang:", v)

	// atomic: เหมาะกับตัวเลขเดี่ยวๆ เช่น counter
	var ops atomic.Int64
	for range 50 {
		wg.Go(func() {
			for range 100 {
				ops.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println("ops:", ops.Load())

	// sync.Once: ทำครั้งเดียว
	var once sync.Once
	for range 3 {
		once.Do(func() { fmt.Println("init once") })
	}

	// sync.OnceValue: ทำครั้งเดียวและจำผลลัพธ์ไว้
	for range 3 {
		wg.Go(func() { _ = getConfig() })
	}
	wg.Wait()
	fmt.Println("env:", getConfig()["env"])
}
