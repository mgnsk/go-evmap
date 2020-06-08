package evmap

import (
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

// if a single writer (sync) is used, then no mutex needed

// if multi writer (concurrent

// get a reader, i)
// start reading +1 (odd)
// finish reading +1 (even)

// swap the atomic pointer for readers

// refresh - swap the pointers, wait for readers counters to get even (all did at least 1 read from the new pointer)
// write log into old pointer

type Map interface {
	Reader() Reader
	// Write the value so that readers may observe a missing value just after Write returns.
	Write(key, value interface{})
	// WriteSync writes so that after it returns, all readers observe the new value.
	WriteSync(key, value interface{})
	// Sync the map.
	Sync()
}

type Reader interface {
	Read(key interface{}) (value interface{}, exists bool)
	Close() error
}

const (
	idle uint64 = iota
	swapping
)

type datamap map[interface{}]interface{}

type rwmap struct {
	r datamap
	w datamap
}

type readermap map[*reader]uint64

type reader struct {
	epoch *uint64
	rmap  *unsafe.Pointer
}

func (r *reader) Read(key interface{}) (interface{}, bool) {
	atomic.AddUint64(r.epoch, 1)
	rmap := *(*datamap)(atomic.LoadPointer(r.rmap))
	defer atomic.AddUint64(r.epoch, 1)
	val, ok := rmap[key]
	return val, ok
}

func (r *reader) Close() error {
	atomic.StoreUint64(r.epoch, math.MaxUint64)
	return nil
}

type evmap struct {
	mu      sync.RWMutex
	epoches []*uint64
	history map[*uint64]uint64
	log     datamap

	r *unsafe.Pointer
	w *unsafe.Pointer

	state uint64
}

func New() Map {
	r := make(datamap)
	w := make(datamap)
	rPtr := unsafe.Pointer(&r)
	wPtr := unsafe.Pointer(&w)
	return &evmap{
		history: make(map[*uint64]uint64),
		r:       &rPtr,
		w:       &wPtr,
		log:     make(datamap),
	}
}

func (m *evmap) Reader() Reader {
	m.mu.Lock()
	defer m.mu.Unlock()

	epoch := new(uint64)
	m.epoches = append(m.epoches, epoch)
	rd := &reader{
		epoch: epoch,
		rmap:  m.r,
	}
	return rd
}

func (m *evmap) Write(key, value interface{}) {
	panic("unimplemented")
	//m.mu.Lock()
	//defer m.mu.Unlock()
	//w := *(*datamap)(atomic.LoadPointer(m.w))
	//w[key] = value
	//m.log[key] = value
}

func (m *evmap) WriteSync(key, value interface{}) {
	defer m.Sync()

	m.mu.Lock()
	w := *(*datamap)(atomic.LoadPointer(m.w))
	w[key] = value
	m.log[key] = value
	m.mu.Unlock()

}

func (m *evmap) Sync() {
	var newW datamap
	if atomic.CompareAndSwapUint64(&m.state, idle, swapping) {
		defer atomic.StoreUint64(&m.state, idle)

		m.mu.Lock()
		newW = *(*datamap)(atomic.SwapPointer(m.r, atomic.SwapPointer(m.w, *m.r)))
		m.mu.Unlock()

		m.wait()

		m.mu.Lock()
		defer m.mu.Unlock()
		for k, v := range m.log {
			newW[k] = v
			delete(m.log, k)
		}

		return
	}

	// Wait until a concurrent sync is finished.
	for {
		m.mu.Lock()
		m.mu.Unlock()
		if atomic.LoadUint64(&m.state) == idle {
			return
		}
	}
}

func (m *evmap) wait() {
	// todo clean up
	done := uint64(0)
	seen := make(map[*uint64]struct{})
	for {
		m.mu.Lock()
		func() {
			if len(m.epoches) == 0 {
				atomic.StoreUint64(&done, 1)
				return
			}
			for i, x := range m.epoches {
				if _, ok := seen[x]; ok {
					if len(seen) == len(m.epoches) {
						// All readers are finished accessing the data referenced
						// by their current read pointer and will only access
						// the swapped pointer from now on.
						atomic.StoreUint64(&done, 1)
						return
					}
					continue
				}
				epoch := atomic.LoadUint64(x)
				if epoch == math.MaxUint64 {
					delete(m.history, x)
					delete(seen, x)
					m.epoches[i] = nil
					if len(m.epoches) == 1 {
						// It was the last reader.
						atomic.StoreUint64(&done, 1)
						return
					}
				} else if epoch == 0 {
					// The reader has not read any pointer yet.
					seen[x] = struct{}{}
				} else if epoch%2 == 0 && epoch > m.history[x] {
					// Reader has seen the new pointer after last swap.
					seen[x] = struct{}{}
					m.history[x] = epoch
				}
			}
		}()
		alive := m.epoches[:0]
		for _, x := range m.epoches {
			if x != nil {
				alive = append(alive, x)
			}
		}
		m.epoches = alive
		m.mu.Unlock()

		if atomic.LoadUint64(&done) == 1 {
			return
		}
	}
}
