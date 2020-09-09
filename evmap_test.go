package evmap

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestMap(t *testing.T) {
	read := func(r Reader, key string, expectExists bool) {
		val, exists := r.Load(key)
		if exists != expectExists {
			panic(fmt.Sprintln("invalid expect value", exists, "expected:", expectExists))
		}
		if exists && val != "value" {
			panic(fmt.Sprintln("wrong value", val))
		}
	}

	m := New(0)

	// Take a reader to the map.
	r := m.Reader()
	defer r.Close()

	// Write some values
	m.Store("key", "value")
	m.Store("key2", "value")

	read(r, "key", true)
	read(r, "key2", true)
}

// Number of unique keys.
const n = 100

func genKeys() []int {
	keys := make([]int, n)
	for i := 0; i < n; i++ {
		keys[i] = i
	}
	return keys
}

var (
	backgroundConcurrency = runtime.GOMAXPROCS(0)
	keys                  = genKeys()
	total                 uint64
)

func nextKey() int {
	return keys[atomic.AddUint64(&total, 1)%n]
}

type benchmark func(b *testing.B)

func BenchmarkEvMapReadSync(b *testing.B) {
	store := func(m Map) func() {
		return func() {
			m.Store(nextKey(), "value")
		}
	}

	b.Run("read", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(m Map) func()
			bench         func(m Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         evMapSingleRead,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         evMapSingleRead,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         evMapMultiRead,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         evMapMultiRead,
			},
		} {
			m := New(0)
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(m), bc.bench(m), goBackground)
			m = nil
		}
	})
}

func BenchmarkEvMapRead100Millisecond(b *testing.B) {
	store := func(m Map) func() {
		return func() {
			m.Store(nextKey(), "value")
		}
	}

	b.Run("read", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(m Map) func()
			bench         func(m Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         evMapSingleRead,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         evMapSingleRead,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         evMapMultiRead,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         evMapMultiRead,
			},
		} {
			m := New(100 * time.Millisecond)
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(m), bc.bench(m), goBackground)
			m = nil
		}
	})
}

func BenchmarkMutexMapRead(b *testing.B) {
	store := func(mu sync.Locker, m datamap) func() {
		return func() {
			mu.Lock()
			m[nextKey()] = "value"
			mu.Unlock()
		}
	}

	b.Run("read", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(sync.Locker, datamap) func()
			bench         func(sync.Locker, datamap) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         mutexMapSingleRead,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         mutexMapSingleRead,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         mutexMapMultiRead,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         mutexMapMultiRead,
			},
		} {
			m := make(datamap)
			var mu sync.RWMutex
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(&mu, m), bc.bench(mu.RLocker(), m), goBackground)
			m = nil
		}
	})
}

func BenchmarkSyncMapRead(b *testing.B) {
	store := func(m *sync.Map) func() {
		return func() {
			m.Store(nextKey(), "value")
		}
	}

	b.Run("read", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(*sync.Map) func()
			bench         func(*sync.Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         syncMapSingleRead,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         syncMapSingleRead,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: 1,
				bg:            store,
				bench:         syncMapMultiRead,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            store,
				bench:         syncMapMultiRead,
			},
		} {
			var m sync.Map
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(&m), bc.bench(&m), goBackground)
		}
	})
}

func BenchmarkEvMapWriteSync(b *testing.B) {
	b.Run("write", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bench         func(m Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bench:         evMapSingleWrite,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: 1,
				bench:         evMapMultiWrite,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bench:         evMapSingleWrite,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bench:         evMapMultiWrite,
			},
		} {
			m := New(0)
			// TODO a mess of trying to fit a generic test
			readers := make([]Reader, bc.bgConcurrency)
			for i := 0; i < bc.bgConcurrency; i++ {
				readers[i] = m.Reader()
			}

			runBench(b, bc.name, bc.bgConcurrency, nil, bc.bench(m), func(ctx context.Context, _ int, done, _ func()) {
				for _, r := range readers {
					r := r
					goBackground(ctx, 1, done, func() {
						res, _ := r.Load(nextKey())
						_ = res
					})
				}
			})
			m = nil
		}
	})
}

func BenchmarkEvMapWrite100Millisecond(b *testing.B) {
	b.Run("write", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bench         func(m Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bench:         evMapSingleWrite,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: 1,
				bench:         evMapMultiWrite,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bench:         evMapSingleWrite,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bench:         evMapMultiWrite,
			},
		} {
			m := New(100 * time.Millisecond)
			// TODO a mess of trying to fit a generic test
			readers := make([]Reader, bc.bgConcurrency)
			for i := 0; i < bc.bgConcurrency; i++ {
				readers[i] = m.Reader()
			}

			runBench(b, bc.name, bc.bgConcurrency, nil, bc.bench(m), func(ctx context.Context, _ int, done, _ func()) {
				for _, r := range readers {
					r := r
					goBackground(ctx, 1, done, func() {
						res, _ := r.Load(nextKey())
						_ = res
					})
				}
			})
			m = nil
		}
	})
}

func BenchmarkMutexMapWrite(b *testing.B) {
	load := func(mu sync.Locker, m datamap) func() {
		return func() {
			mu.Lock()
			val, _ := m[nextKey()]
			mu.Unlock()
			_ = val
		}
	}

	b.Run("write", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(sync.Locker, datamap) func()
			bench         func(sync.Locker, datamap) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            load,
				bench:         mutexMapSingleWrite,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: 1,
				bg:            load,
				bench:         mutexMapMultiWrite,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            load,
				bench:         mutexMapSingleWrite,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            load,
				bench:         mutexMapMultiWrite,
			},
		} {
			m := make(datamap)
			var mu sync.RWMutex
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(mu.RLocker(), m), bc.bench(&mu, m), goBackground)
			m = nil
		}
	})
}

func BenchmarkSyncMapWrite(b *testing.B) {
	load := func(m *sync.Map) func() {
		return func() {
			val, _ := m.Load(nextKey())
			_ = val
		}
	}

	b.Run("write", func(b *testing.B) {
		for _, bc := range []struct {
			name          string
			bgConcurrency int
			bg            func(*sync.Map) func()
			bench         func(*sync.Map) benchmark
		}{
			{
				name:          "single writer single reader",
				bgConcurrency: 1,
				bg:            load,
				bench:         syncMapSingleWrite,
			},
			{
				name:          "multi writer single reader",
				bgConcurrency: 1,
				bg:            load,
				bench:         syncMapMultiWrite,
			},
			{
				name:          "single writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            load,
				bench:         syncMapSingleWrite,
			},
			{
				name:          "multi writer multi reader",
				bgConcurrency: backgroundConcurrency,
				bg:            load,
				bench:         syncMapMultiWrite,
			},
		} {
			var m sync.Map
			runBench(b, bc.name, bc.bgConcurrency, bc.bg(&m), bc.bench(&m), goBackground)
		}
	})
}

func goBackground(ctx context.Context, concurrency int, done, cb func()) {
	for i := 0; i < concurrency; i++ {
		go func() {
			defer done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					cb()
				}
			}
		}()
	}
}

func runBench(b *testing.B, name string, bgConcurrency int, bg func(), bench func(*testing.B), g func(context.Context, int, func(), func())) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(bgConcurrency)
	g(ctx, bgConcurrency, wg.Done, bg)

	b.Run(name, bench)

	cancel()
	wg.Wait()

	runtime.GC()
}

func evMapSingleRead(m Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		r := m.Reader()
		defer r.Close()
		var res interface{}
		var exists bool
		for i := 0; i < b.N; i++ {
			res, exists = r.Load(nextKey())
		}
		_, _ = res, exists
	}
}

func evMapMultiRead(m Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			r := m.Reader()
			defer r.Close()
			var res interface{}
			var exists bool
			for pb.Next() {
				res, exists = r.Load(nextKey())
			}
			_, _ = res, exists
		})
	}
}

func mutexMapSingleRead(mu sync.Locker, m datamap) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		var res interface{}
		var exists bool
		for i := 0; i < b.N; i++ {
			mu.Lock()
			res, exists = m[nextKey()]
			mu.Unlock()
		}
		_, _ = res, exists
	}
}

func mutexMapMultiRead(mu sync.Locker, m datamap) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var res interface{}
			var exists bool
			for pb.Next() {
				mu.Lock()
				res, exists = m[nextKey()]
				mu.Unlock()
			}
			_, _ = res, exists
		})
	}
}

func syncMapSingleRead(m *sync.Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		var res interface{}
		var exists bool
		for i := 0; i < b.N; i++ {
			res, exists = m.Load(nextKey())
		}
		_, _ = res, exists
	}
}

func syncMapMultiRead(m *sync.Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var res interface{}
			var exists bool
			for pb.Next() {
				res, exists = m.Load(nextKey())
			}
			_, _ = res, exists
		})
	}
}

func evMapSingleWrite(m Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Store(nextKey(), "value")
		}
	}
}

func evMapMultiWrite(m Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Store(nextKey(), "value")
			}
		})
	}
}

func mutexMapSingleWrite(mu sync.Locker, m datamap) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mu.Lock()
			m[nextKey()] = "value"
			mu.Unlock()
		}
	}
}

func mutexMapMultiWrite(mu sync.Locker, m datamap) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				m[nextKey()] = "value"
				mu.Unlock()
			}
		})
	}
}

func syncMapSingleWrite(m *sync.Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Store(nextKey(), "value")
		}
	}
}

func syncMapMultiWrite(m *sync.Map) benchmark {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Store(nextKey(), "value")
			}
		})
	}
}
