package evmap

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Map is a lock-free eventually consistent map.
// Writes do not block reads but block other writes.
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

type evmap struct {
	refreshRate time.Duration
	initRefresh sync.Once
	state       uint64
	readers     sync.Map
	lmap        *unsafe.Pointer
	rmap        *unsafe.Pointer
	log         datamap
	mu          sync.Mutex
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
	m.mu.Lock()
	defer m.mu.Unlock()

	writeMap := *(*datamap)(atomic.LoadPointer(m.rmap))
	writeMap[key] = value
	m.log[key] = value

	if m.refreshRate == 0 {
		m.refresh()
	} else {
		m.initRefresh.Do(func() {
			go m.startRefresh(context.TODO())
		})
	}
}

func (m *evmap) startRefresh(ctx context.Context) {
	ticker := time.NewTicker(m.refreshRate)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			m.refresh()
			m.mu.Unlock()
		}
	}
}

func (m *evmap) refresh() {
	newRight := *(*datamap)(atomic.SwapPointer(m.lmap, atomic.SwapPointer(m.rmap, *m.lmap)))
	m.wait()
	for k, v := range m.log {
		newRight[k] = v
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
