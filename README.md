# go-weakref

A generic weak reference in Go. Inspired by https://github.com/ivanrad/go-weakref

It is (naively) tested against race conditions.

# Usage

```go
package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/szmcdull/go-weakref"
)

func main() {
	v := 123
	p := weakref.NewWeakRef(&v)
	fmt.Println(weakref.IsAlive(p)) // true
	fmt.Println(*weakref.Get(p))    // 123
	runtime.KeepAlive(v)

	runtime.GC()
	time.Sleep(time.Millisecond)
	runtime.GC()
	time.Sleep(time.Millisecond)

	fmt.Println(weakref.IsAlive(p)) // false
}
```