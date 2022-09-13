package weakref

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

type Int struct {
	*int
}

var (
	wg          sync.WaitGroup
	testingRace bool
)

func makeRef() *WeakRef[int] {
	a := 123
	p := &a
	r := NewWeakRef(p)
	return r
}

func testNewWeakRef(i int, t *testing.T) {
	//r := makeRef()
	a := 123
	p := &a
	r := NewWeakRef(p)

	if !IsAlive(r) {
		t.Error(`early freed`)
	}
	if *Get(r) != 123 {
		t.Fail()
	}
	runtime.KeepAlive(a)
	_ = &a // keep a in memory til here

	runtime.GC()
	time.Sleep(time.Millisecond * 10)
	runtime.GC()
	time.Sleep(time.Millisecond * 10)

	p = Get(r)
	isAlive := IsAlive(r)
	if p != nil && *p != 123 {
		t.Error(`bad value`)
	}
	if p == nil && isAlive {
		t.Errorf(`wrong status %p, %t`, p, isAlive)
	}
	// when p is not null isAlive may be false

	if testingRace {
		wg.Done()
	}
}

func testNewFromSlice(i int, t *testing.T) {
	a := []int{123, 222, 333}
	r := NewWeakRef(&a[0])
	if !IsAlive(r) {
		t.Error(`early freed`)
	}
	if *Get(r) != 123 {
		t.Fail()
	}

	a = append(a, make([]int, 2560)...)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	p := Get(r)
	if p != nil {
		if !IsAlive(r) {
			t.Error(`wrong status`)
		}
		if *p != 123 {
			t.Error(`bad pointer`)
		}
	} else {
		if IsAlive(r) {
			t.Error(`wrong status`)
		}
	}

	// wrap a defer function to test if pointer is invalid
	func() {
		p := Get(r)
		// if p != nil {
		// 	if *p == a[0] {
		// 		t.Error(`slice not moved`)
		// 	}
		// }
		if p != nil {
			if *p != a[0] {
				t.Error(`bad pointer`)
			}
		}

	}()

	if testingRace {
		wg.Done()
	}
}

func TestOnce(t *testing.T) {
	testNewWeakRef(-1, t)
	testNewFromSlice(-1, t)
}

func TestRace(t *testing.T) {
	testingRace = true
	testCount := 500000
	wg.Add(testCount * 2)
	for i := 0; i < testCount; i++ {
		ii := i
		go testNewWeakRef(ii, t)
		go testNewFromSlice(ii, t)
	}
	wg.Wait()
	testingRace = false
}
