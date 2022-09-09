package weakref

import (
	"runtime"
	"testing"
	"time"
)

func makeRef() *WeakRef[int] {
	a := 123
	p := &a
	r := NewWeakRef(p)
	return r
}

func TestNewWeakRef(t *testing.T) {
	r := makeRef()
	if !IsAlive(r) {
		t.Error(`early freed`)
	}
	if *Get(r) != 123 {
		t.Fail()
	}
	runtime.GC()
	time.Sleep(time.Millisecond * 100)
	runtime.GC()
	time.Sleep(time.Millisecond * 100)
	if IsAlive(r) {
		t.Error(`not freed`)
	}
}
