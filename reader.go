package evmap

import (
	"math"
	"sync/atomic"
	"unsafe"
)

type reader struct {
	epoch   *uint64
	history *uint64
	lmap    *unsafe.Pointer
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
	readMap := *(*datamap)(atomic.LoadPointer(r.lmap))
	value, ok = readMap[key]
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
