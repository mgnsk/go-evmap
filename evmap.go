package evmap

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Map is a lock-free eventually consistent map.
type Map interface {
	Reader() Reader
	Store(key, value interface{})
}

// Reader is a map reader.
type Reader interface {
	Load(key interface{}) (value interface{}, exists bool)
	Close() error
}

type datamap map[interface{}]interface{}

type reader struct {
	epoch   *uint64
	history *uint64
	flipped *uint64
	rwmap   unsafe.Pointer
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

func (r *reader) wait() (exists bool) {
	for {
		epoch := atomic.LoadUint64(r.epoch)
		history := atomic.LoadUint64(r.history)
		if epoch == math.MaxUint64 {
			return false
		} else if epoch == 0 {
			// The reader has not read any pointer yet.
			return true
		} else if epoch%2 == 0 && epoch > history {
			// Reader has seen the new pointer after last swap.
			atomic.StoreUint64(r.history, epoch)
			return true
		}
	}
}

const (
	idle uint64 = iota
	writing
)

var counter uint64

type evmap struct {
	refreshRate time.Duration
	readers     sync.Map
	flipped     *uint64
	log         datamap

	rmap datamap
	wmap datamap

	state       uint64
	lastRefresh time.Time
}

// New creates an empty concurrent map.
func New(refreshRate time.Duration) Map {
	return &evmap{
		refreshRate: refreshRate,
		flipped:     new(uint64),
		rmap:        make(datamap),
		wmap:        make(datamap),
		log:         make(datamap),
	}
}

// Reader registers a reader for the map.
// The reader must be closed when no longer in use.
func (m *evmap) Reader() Reader {
	rd := &reader{
		epoch:   new(uint64),
		history: new(uint64),
		flipped: m.flipped,
		rmap:    m.rmap,
		wmap:    m.wmap,
	}

	m.readers.Store(rd, nil)

	return rd
}

// Store a value into the map. After Store returns,
// all readers observe the new value.
func (m *evmap) Store(key, value interface{}) {
	defer atomic.AddUint64(&counter, 1)

	for {
		if !atomic.CompareAndSwapUint64(&m.state, idle, writing) {
			continue
		}

		defer atomic.StoreUint64(&m.state, idle)

		if m.refreshRate == 0 || time.Since(m.lastRefresh) > m.refreshRate {
			defer func() {
				m.swap()
				m.lastRefresh = time.Now()
			}()
		}

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

func (m *evmap) wait() {
	m.readers.Range(func(rd, _ interface{}) bool {
		if !rd.(*reader).wait() {
			m.readers.Delete(rd)
		}
		return true
	})
}
