package evmap

import (
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

const (
	idle uint64 = iota
	writing
)

type evmap struct {
	refreshRate time.Duration
	lastRefresh time.Time
	state       uint64
	readers     sync.Map
	lmap        *unsafe.Pointer
	rmap        *unsafe.Pointer
	log         datamap
}

// New creates an empty concurrent map.
func New(refreshRate time.Duration) Map {
	readmap := make(datamap)
	writemap := make(datamap)
	lmap := unsafe.Pointer(&readmap)
	rmap := unsafe.Pointer(&writemap)
	return &evmap{
		refreshRate: refreshRate,
		lmap:        &lmap,
		rmap:        &rmap,
		log:         make(datamap),
	}
}

// Reader registers a reader for the map.
// The reader must be closed when no longer in use.
func (m *evmap) Reader() Reader {
	rd := &reader{
		epoch:   new(uint64),
		history: new(uint64),
		lmap:    m.lmap,
	}
	m.readers.Store(rd, nil)
	return rd
}

// Store a value into the map.
func (m *evmap) Store(key, value interface{}) {
	for {
		if !atomic.CompareAndSwapUint64(&m.state, idle, writing) {
			continue
		}
		defer atomic.StoreUint64(&m.state, idle)
		if m.refreshRate == 0 {
			defer m.swap()
		} else if now := time.Now(); now.Sub(m.lastRefresh) > m.refreshRate {
			defer m.swap()
			m.lastRefresh = now
		}
		writeMap := *(*datamap)(atomic.LoadPointer(m.rmap))
		writeMap[key] = value
		m.log[key] = value
		return
	}
}

func (m *evmap) swap() {
	writeMap := *(*datamap)(atomic.SwapPointer(m.lmap, atomic.SwapPointer(m.rmap, *m.lmap)))
	m.wait()
	for k, v := range m.log {
		writeMap[k] = v
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
