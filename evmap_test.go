package evmap

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	read := func(r Reader, key string, expectExists bool) {
		val, exists := r.Read(key)
		if exists != expectExists {
			panic(fmt.Sprintln("invalid expect value", exists, "expected:", expectExists))
		}
		if exists && val != "value" {
			panic(fmt.Sprintln("wrong value", val))
		}
	}

	m := New()

	// Take a reader to the map.
	r := m.Reader()
	defer r.Close()

	// Write some values
	m.WriteSync("key", "value")
	m.WriteSync("key2", "value")

	read(r, "key", true)
	read(r, "key2", true)
}

const writeEveryNRead = 1

const numWriters = 4

// Number of unique keys.
const n = 1000

func genKeys() []int {
	keys := make([]int, n)
	for i := 0; i < n; i++ {
		keys[i] = i
	}
	return keys
}

func BenchmarkEvMap(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	m := New()

	totalReads := uint64(0)
	keys := genKeys()

	for i := 0; i < numWriters; i++ {
		go func() {
			for {
				if total := atomic.LoadUint64(&totalReads); total%writeEveryNRead == 0 {
					key := keys[total%n]
					m.WriteSync(key, "value")
				}
			}
		}()
	}

	b.RunParallel(func(pb *testing.PB) {
		r := m.Reader()
		defer r.Close()
		var res interface{}
		var exists bool
		for pb.Next() {
			key := keys[atomic.AddUint64(&totalReads, 1)%n]
			res, exists = r.Read(key)
		}
		_, _ = res, exists
	})
}

func BenchmarkMutexMap(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	mu := sync.RWMutex{}
	m := make(datamap)

	totalReads := uint64(0)
	keys := genKeys()

	for i := 0; i < numWriters; i++ {
		go func() {
			for {
				if total := atomic.LoadUint64(&totalReads); total%writeEveryNRead == 0 {
					key := keys[total%n]
					mu.Lock()
					m[key] = "value"
					mu.Unlock()
				}
			}
		}()
	}

	b.RunParallel(func(pb *testing.PB) {
		var res interface{}
		var exists bool
		for pb.Next() {
			key := keys[atomic.AddUint64(&totalReads, 1)%n]
			mu.RLock()
			res, exists = m[key]
			mu.RUnlock()
		}
		_, _ = res, exists
	})
}

func BenchmarkSyncMap(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	m := sync.Map{}

	totalReads := uint64(0)
	keys := genKeys()

	for i := 0; i < numWriters; i++ {
		go func() {
			for {
				if total := atomic.LoadUint64(&totalReads); total%writeEveryNRead == 0 {
					key := keys[total%n]
					m.Store(key, "value")
				}
			}
		}()
	}

	b.RunParallel(func(pb *testing.PB) {
		var res interface{}
		var exists bool
		for pb.Next() {
			key := keys[atomic.AddUint64(&totalReads, 1)%n]
			res, exists = m.Load(key)
		}
		_, _ = res, exists
	})
}
