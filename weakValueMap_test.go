package weakref

import (
	"runtime"
	"testing"
	"time"
)

func TestWeakValueMap(t *testing.T) {
	m := NewWeakValueMap[int, []int]()
	v := make([]int, 10000)
	v[0] = 123
	m.Store(1, &v)
	if r, ok := m.Load(1); ok {
		if r[0] != 123 {
			t.Errorf(`expected 123 got %d`, r[0])
		}
	} else {
		t.Error(`v is gone`)
	}
	runtime.KeepAlive(v)
	runtime.GC()
	time.Sleep(time.Millisecond)
	runtime.GC()
	time.Sleep(time.Millisecond)
	if _, ok := m.Load(1); ok {
		t.Error(`v is not garbage collected`)
	}
}
