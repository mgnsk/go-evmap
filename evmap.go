package evmap

import (
	"math"
	"sync"
	"sync/atomic"
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
	epoch   *uint64
	flipped *uint64
	rmap    datamap
	wmap    datamap
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (r *reader) Load(key interface{}) (value interface{}, ok bool) {
	atomic.CompareAndSwapUint64(r.epoch, math.MaxUint64-1, 0)
	if atomic.AddUint64(r.epoch, 1)%2 == 0 {
		panic("evmap: reader used concurrently")
	}
	defer atomic.AddUint64(r.epoch, 1)
	switch atomic.LoadUint64(r.flipped) {
	case 0:
		value, ok = r.rmap[key]
	default:
		value, ok = r.wmap[key]
	}
	return
}

// Close unregisters the reader.
func (r *reader) Close() error {
	atomic.StoreUint64(r.epoch, math.MaxUint64)
	return nil
}

const (
	idle uint64 = iota
	swapping
	writing
)

type evmap struct {
	mu      sync.Mutex
	epoches []*uint64
	flipped *uint64
	history map[*uint64]uint64
	log     datamap

	rmap datamap
	wmap datamap

	state uint64
}

// New creates an empty concurrent map.
func New() Map {
	return &evmap{
		flipped: new(uint64),
		history: make(map[*uint64]uint64),
		rmap:    make(datamap),
		wmap:    make(datamap),
		log:     make(datamap),
	}
}

// Reader registers a reader for the map.
// The reader must be closed when no longer in use.
func (m *evmap) Reader() Reader {
	m.mu.Lock()
	defer m.mu.Unlock()

	epoch := new(uint64)
	m.epoches = append(m.epoches, epoch)
	rd := &reader{
		epoch:   epoch,
		flipped: m.flipped,
		rmap:    m.rmap,
		wmap:    m.wmap,
	}
	return rd
}

// Store a value into the map. After Store returns,
// all readers observe the new value.
func (m *evmap) Store(key, value interface{}) {
	defer m.swap()

	for {
		if !atomic.CompareAndSwapUint64(&m.state, idle, writing) {
			continue
		}

		defer atomic.StoreUint64(&m.state, idle)

		switch atomic.LoadUint64(m.flipped) {
		case 0:
			m.wmap[key] = value
		default:
			m.rmap[key] = value
		}
		m.log[key] = value

		return
	}
}

func (m *evmap) swap() {
	for {
		if !atomic.CompareAndSwapUint64(&m.state, idle, swapping) {
			continue
		}
		defer atomic.StoreUint64(&m.state, idle)

		flipped := 0

		if !atomic.CompareAndSwapUint64(m.flipped, 0, 1) {
			atomic.StoreUint64(m.flipped, 0)
		} else {
			flipped = 1
		}

		m.wait()

		for k, v := range m.log {
			switch flipped {
			case 0:
				m.wmap[k] = v
			case 1:
				m.rmap[k] = v
			}
			delete(m.log, k)
		}

		return
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
				if x == nil {
					continue
				}
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
