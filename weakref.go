package weakref

import (
	"fmt"
	"runtime"
	"unsafe"
)

type (
	WeakRef[T any] struct {
		p uintptr
	}
)

// Caution should be taken when v from a slice.
// When the slice grows its items would be copied to a new address,
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
	r.p = 0
}

func IsAlive[T any](r *WeakRef[T]) bool {
	return r.p != 0
}

func Get[T any](r *WeakRef[T]) (result *T) {
	defer func() {
		if e := recover(); e != nil {
			s := fmt.Sprintf(`%v`, e)
			if s == `runtime error: invalid memory address or nil pointer dereference` { // finalizer not called yet, but invalid pointer detected
				r.p = 0
				result = nil
			} else {
				panic(e)
			}
		}
	}()

	result = (*T)(unsafe.Pointer(r.p))

	// when the finalizer is not called soon enough, invalid pointer may be used.
	// test if the pointer is still valid
	_ = *result

	return result
}
