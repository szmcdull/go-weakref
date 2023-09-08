package weakref

import (
	"github.com/szmcdull/glinq/gmap"
)

type (
	WeakValueMap[K comparable, V any] struct {
		m *gmap.SyncMap[K, *WeakRef[V]]
	}
)

func NewWeakValueMap[K comparable, V any]() *WeakValueMap[K, V] {
	return &WeakValueMap[K, V]{
		m: gmap.NewSyncMap[K, *WeakRef[V]](),
	}
}

func (me *WeakValueMap[K, V]) Store(k K, v *V) {
	w := NewWeakRef(v)
	w.SetCallback(func(v *V) bool {
		me.m.Delete(k)
		return true
	})
	me.m.Store(k, w)
}

func (me *WeakValueMap[K, V]) Load(k K) (result V, ok bool) {
	if r, ok2 := me.m.Load(k); ok2 {
		p := Get(r)
		if p != nil {
			result = *p
			ok = true
		} else {
			ok = false
		}
	}
	return
}
