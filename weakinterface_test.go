package weakref

import (
	"fmt"
	"runtime"
	"sync/atomic"
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

	if !IsAliveI(r) {
		t.Error(`early freed`)
	}
	w := GetInterface(r)
	if w.A() != 123 {
		t.Fail()
	}
	time.Sleep(time.Millisecond * 10)
	if w.A() != 123 {
		t.Fail()
	}
	time.Sleep(time.Second)
	if w.A() != 123 {
		t.Fail()
	}
	runtime.KeepAlive(ps)

	_ = ps // keep a in memory til here
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)

	if IsAliveI(r) {
		if !testingRace { // finalizer is called in a separated GoProc and may not finish yet in race condition
			t.Error(`not freed`)
		}
	} else {
		atomic.AddInt64(&interfaceFinalizeCount, 1)
	}

	if testingRace {
		wg.Done()
	}
}

func testNewInterfaceFromSlice(i int, t *testing.T) {
	a := []TestInterface{&PointerStruct{123}, &PointerStruct{222}, &PointerStruct{333}}
	r := NewWeakInterface(a[0])
	if !IsAliveI(r) {
		t.Error(`early freed`)
	}
	if GetInterface(r).A() != 123 {
		t.Fail()
	}

	a = append(a, make([]TestInterface, 4096)...)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)

	if !IsAliveI(r) {
		t.Error(`early free`)
	}
	if GetInterface(r) != a[0] {
		t.Error(`bad reference`)
	}

	if testingRace {
		wg.Done()
	}
}

func TestInterfaceOnce(t *testing.T) {
	testNewWeakInterface(-1, t)
	testNewInterfaceFromSlice(-1, t)
}

func TestInterfaceRace(t *testing.T) {
	testingRace = true
	testCount := 500000
	wg.Add(testCount * 2)
	for i := 0; i < testCount; i++ {
		ii := i
		go testNewWeakInterface(ii, t)
		go testNewInterfaceFromSlice(ii, t)
	}
	wg.Wait()
	testingRace = false

	fmt.Println(`WeakInterface test times: `, testCount, `, finalizeCount: `, interfaceFinalizeCount)
	if interfaceFinalizeCount == 0 {
		t.Error(`none finalized`)
	}
}
