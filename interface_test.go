package weakref

import (
	"runtime"
	"testing"
	"time"
)

type (
	TestInterface interface {
		A() int
	}
	Struct struct {
		v int
	}
	PointerStruct struct {
		v int
	}
)

func (s Struct) A() int {
	return s.v
}

func (s *PointerStruct) A() int {
	return s.v
}

func testNewWeakInterface(i int, t *testing.T) {
	//r := makeRef()
	var ps TestInterface
	s := Struct{123}
	ps = &PointerStruct{123}

	func() {
		defer func() {
			if e := recover(); e == nil {
				t.Error(`should not accept non-pointer interface`)
			}
		}()

		NewWeakInterface(s)
	}()

	r := NewWeakInterface(ps)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)

	if !IsAlive(&r.WeakRef) {
		t.Error(`early freed`)
	}
	w := GetInterface(r)
	if w.A() != 123 {
		t.Fail()
	}
	runtime.KeepAlive(ps)
	_ = ps // keep a in memory til here
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	if IsAlive(&r.WeakRef) {
		if !testingRace { // finalizer is called in a separated GoProc and may not finish yet in race condition
			t.Error(`not freed`)
		}
	}

	if testingRace {
		wg.Done()
	}
}

func testNewInterfaceFromSlice(i int, t *testing.T) {
	a := []int{123, 222, 333}
	r := NewWeakInterface(&a[0])
	if !IsAlive(&r.WeakRef) {
		t.Error(`early freed`)
	}
	if *GetInterface(r) != 123 {
		t.Fail()
	}

	a = append(a, make([]int, 256)...)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	if IsAlive(&r.WeakRef) {
		t.Error(`not freed`)
	}

	// wrap a defer function to test if pointer is invalid
	func() {
		defer func() {
			e := recover()
			if e == nil {
				t.Error(`slice not moved`)
			}
		}()
		if *GetInterface(r) == a[0] {
			t.Error(`slice not moved`)
		}
	}()

	if testingRace {
		wg.Done()
	}
}

func TestInterfaceOnce(t *testing.T) {
	testNewWeakInterface(-1, t)
	//testNewInterfaceFromSlice(-1, t)
}

func TestInterfaceRace(t *testing.T) {
	testingRace = true
	testCount := 500000
	wg.Add(testCount * 1)
	for i := 0; i < testCount; i++ {
		ii := i
		go testNewWeakInterface(ii, t)
		//go testNewInterfaceFromSlice(ii, t)
	}
	wg.Wait()
	testingRace = false
}
