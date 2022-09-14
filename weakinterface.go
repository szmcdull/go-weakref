package weakref

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

type (
	WeakInterface[T any] struct {
		GoInterface
	}

	GoInterface struct {
		typ  uintptr //*rtype
		word uintptr
	}
)

// safe with slice of interface.
// because interface itself is already a reference to some data.
// it is alright that the interface itself is garbage collected as long as the data pointed to is alive.
func NewWeakInterface[T any](v T) *WeakInterface[T] {
	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Pointer {
		panic(`NewWeakInterface from a non-pointer interface`)
	}

	intf := (*GoInterface)(unsafe.Pointer(&v))
	sz := unsafe.Sizeof(v)
	sz2 := unsafe.Sizeof(*intf)
	if sz2 != sz {
		panic(`NewWeakInterface expected a interface instance`)
	}

	result := &WeakInterface[T]{
		GoInterface: *intf,
	}

	runtime.SetFinalizer(v, func(v T) { _OnInterfaceDataFinalized(result) })

	return result
}

func _OnInterfaceDataFinalized[T any](w *WeakInterface[T]) {
	w.word = 0
	w.typ = 0
}

func IsAliveI[T any](w *WeakInterface[T]) bool {
	return w.word != 0
}

func GetInterface[T any](r *WeakInterface[T]) (result T) {
	defer func() {
		if e := recover(); e != nil {
			s := fmt.Sprintf(`%v`, e)
			if s == `runtime error: invalid memory address or nil pointer dereference` { // finalizer not called yet, but invalid pointer detected
				r.word = 0
				r.typ = 0
				var a T
				result = a
			} else {
				panic(e)
			}
		}
	}()

	intf := (*GoInterface)(unsafe.Pointer(&result))
	*intf = r.GoInterface

	// when the finalizer is not called soon enough, invalid pointer may be used.
	// test if the pointer is still valid
	_ = *(*int)(unsafe.Pointer(intf.word))

	return result
}
