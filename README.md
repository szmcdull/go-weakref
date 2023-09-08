# go-weakref

A generic weak reference in Go. Inspired by https://github.com/ivanrad/go-weakref

It is (naively) tested against race conditions.

# Usage

```go
package main

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/szmcdull/go-weakref"
)

func main() {
	v := 123
	wr := weakref.NewWeakRef(&v)
	fmt.Println(weakref.IsAlive(wr)) // true
	fmt.Println(*weakref.Get(wr))    // 123
	runtime.KeepAlive(v)

	runtime.GC()
	time.Sleep(time.Millisecond)
	runtime.GC()
	time.Sleep(time.Millisecond)

	fmt.Println(weakref.IsAlive(wr)) // false

	err := errors.New(`456`)
	wi := weakref.NewWeakInterface(err)
	fmt.Println(weakref.GetInterface(wi).Error()) // 456
	runtime.KeepAlive(err)

	//weakref.NewWeakRef(errors.New(`123`)) // compiler error: NewWeakRef expects a pointer argument
	//weakref.NewWeakInterface(123) 		// panic, 123 is not a interface
	//weakref.NewWeakInterface(any(123)) 	// panic, 123 is a non-pointer interface
}
```