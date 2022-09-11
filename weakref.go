package weakref

import (
	"runtime"
	"unsafe"
)

type (
	WeakRef[T any] struct {
		//sync.Mutex
		// because p is checked in Get() after converted to typed pointer, it is no longer necessary to lock

		p uintptr
	}
)

// Caution should be taken when v from a slice/map.
// When the slice/map grows its items would be copied to a new address,
// the original items are freed, and IsAlive() will return false.
func NewWeakRef[T any](v *T) *WeakRef[T] {
	// we can assume the variable pointed by v is in the heap already,
	// because its address is taken and passed here, even further to SetFinalizer.
	// currently Go does not move an variable if it is already in the heap,
	// so we can just save v to a uintptr and safely use it later.
	// but this behavior may change in future versions.

	result := &WeakRef[T]{
		p: uintptr(unsafe.Pointer(v)),
	}
	runtime.SetFinalizer(v, func(v *T) { _OnFinalized(result) })

	return result
}

func _OnFinalized[T any](r *WeakRef[T]) {
	// r.Lock()
	// defer r.Unlock()
	r.p = 0
}

func IsAlive[T any](r *WeakRef[T]) bool {
	return r.p != 0
}

func Get[T any](r *WeakRef[T]) (result *T) {
	// r.Lock()
	// defer r.Unlock()
	defer func() {
		if e := recover(); e != nil { // finalizer not called yet, but invalid pointer detected
			r.p = 0
			result = nil
		}
	}()

	result = (*T)(unsafe.Pointer(r.p))

	// when the finalizer is not called soon enough, invalid pointer may be used.
	// test if the pointer is still valid
	_ = *result

	return result
}
