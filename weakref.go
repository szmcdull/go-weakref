package weakref

import (
	"runtime"
	"sync"
	"unsafe"
)

type (
	WeakRef[T any] struct {
		sync.Mutex
		p uintptr
	}
)

// Caution should be taken when v is an address taken of an item from a slice/map.
// When the slice/map grows the item would be copied to a new address,
// the original one is freed, and IsAlive() will return false.
func NewWeakRef[T any](v *T) *WeakRef[T] {
	result := &WeakRef[T]{
		p: uintptr(unsafe.Pointer(v)),
	}
	runtime.SetFinalizer(v, func(v *T) { _OnFinalized(result) })

	return result
}

func _OnFinalized[T any](r *WeakRef[T]) {
	r.Lock()
	defer r.Unlock()
	r.p = 0
}

func IsAlive[T any](r *WeakRef[T]) bool {
	return r.p != 0
}

func Get[T any](r *WeakRef[T]) (result *T) {
	r.Lock()
	defer r.Unlock()
	defer func() {
		if e := recover(); e != nil { // finalizer not called yet, but invalid pointer detected
			r.p = 0
			result = nil
		}
	}()

	// currently Go does not move an variable if it is already in the heap.
	// but no guaranties in the future.
	result = (*T)(unsafe.Pointer(r.p))

	// when the finalizer is not called soon enough, invalid pointer may be used.
	// test if the pointer is still valid
	_ = *result

	return result
}
