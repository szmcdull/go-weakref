package weakref

import (
	"reflect"
	"runtime"
	"unsafe"
)

type (
	WeakRef[T any] struct {
		//sync.Mutex
		// because p is checked in Get() after converted to typed pointer, it is no longer necessary to lock

		p uintptr
	}

	WeakInterface[T any] struct {
		WeakRef[T]
		t Interface
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

func NewWeakInterface[T any](v T) *WeakInterface[T] {
	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Pointer {
		panic(`NewWeakInterface from a non-pointer interface`)
	}
	//val := p.Elem()

	result := &WeakInterface[T]{
		WeakRef: WeakRef[T]{
			p: p.Pointer(),
		},
		t: Interface{},
	}

	s1 := unsafe.Sizeof(v)
	s2 := unsafe.Sizeof(result.t)
	if s2 != s1 {
		panic(`NewWeakInterface expected a interface instance`)
	}

	runtime.SetFinalizer(v, func(v T) { _OnFinalized(&result.WeakRef) })
	i1 := *(*Interface)(unsafe.Pointer(&v))
	i2 := (*Interface)(unsafe.Pointer(&result.t))
	i2.typ = i1.typ

	iii := *(**T)(unsafe.Pointer(&i2))
	_ = iii

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
		if e := recover(); e == `invalid memory address or nil pointer dereference` { // finalizer not called yet, but invalid pointer detected
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

type Interface struct {
	typ  uintptr //*rtype
	word uintptr
}

func GetInterface[T any](r *WeakInterface[T]) (result T) {
	defer func() {
		if e := recover(); e == `invalid memory address or nil pointer dereference` { // finalizer not called yet, but invalid pointer detected
			r.p = 0
			var a T
			result = a
		}
	}()

	reflect.ValueOf(r.t)
	intf := *(*Interface)(unsafe.Pointer(&r.t))
	intf.word = r.p
	result = *(*T)(unsafe.Pointer(&intf))

	// when the finalizer is not called soon enough, invalid pointer may be used.
	// test if the pointer is still valid
	_ = *(*int)(unsafe.Pointer(r.p))

	if r.p == 0 {
		runtime.Version()
	}

	return result
}
