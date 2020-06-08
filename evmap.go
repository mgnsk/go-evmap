package evmap

import (
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Map is a concurrent non-blocking map.
type Map interface {
	Reader() Reader
	// Store writes so that after it returns, all readers observe the new value.
	Store(key, value interface{})
}

// Reader is a map reader.
type Reader interface {
	Load(key interface{}) (value interface{}, exists bool)
	Close() error
}

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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (r *reader) Load(key interface{}) (value interface{}, ok bool) {
	atomic.AddUint64(r.epoch, 1)
	rmap := *(*datamap)(atomic.LoadPointer(r.rmap))
	val, ok := rmap[key]
	if atomic.AddUint64(r.epoch, 1) == math.MaxUint64-1 {
		// Safely roll over to 0 without stopping at math.MaxUint64
		atomic.AddUint64(r.epoch, 2)
	}
	return val, ok
}

// Close unregisters up the reader.
func (r *reader) Close() error {
	atomic.StoreUint64(r.epoch, math.MaxUint64)
	return nil
}

const (
	idle uint64 = iota
	swapping
)

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

func (m *evmap) Store(key, value interface{}) {
	defer m.sync()

	m.mu.Lock()
	w := *(*datamap)(atomic.LoadPointer(m.w))
	w[key] = value
	m.log[key] = value
	m.mu.Unlock()
}

func (m *evmap) sync() {
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
