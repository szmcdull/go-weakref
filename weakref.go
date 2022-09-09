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

func Get[T any](r *WeakRef[T]) *T {
	r.Lock()
	defer r.Unlock()
	return (*T)(unsafe.Pointer(r.p))
}
