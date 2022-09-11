package weakref

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

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
	_ = &a
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	if IsAlive(r) {
		if !testingRace {
			t.Error(`not freed`)
		}
	}

	wg.Done()
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

	a = append(a, make([]int, 256)...)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	runtime.GC()
	time.Sleep(time.Millisecond * 1)
	if IsAlive(r) {
		t.Error(`not freed`)
	}
	func() {
		defer func() {
			e := recover()
			if e == nil {
				t.Error(`slice not moved`)
			}
		}()
		if *Get(r) == a[0] {
			t.Error(`slice not moved`)
		}
	}()

	wg.Done()
}

func TestRace(t *testing.T) {
	testingRace = true
	testCount := 10000
	wg.Add(testCount * 2)
	for i := 0; i < testCount; i++ {
		ii := i
		go testNewWeakRef(ii, t)
		go testNewFromSlice(ii, t)
	}
	wg.Wait()
	testingRace = false
}
