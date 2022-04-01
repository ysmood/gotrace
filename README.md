# Overview

A lib for monitoring runtime goroutine stack.
Such as wait for goroutines to exit, leak detection, etc.

## Features

- `context.Context` first design
- Concurrent leak detection
- No dependencies and 100% test coverage
- Provides handy low-level APIs to extend the lib

## Guides

Check the [examples](examples_test.go) and [godoc](https://pkg.go.dev/github.com/ysmood/gotrace) for detailed usage.

Let's get started with a typical async producer-consumer model, can you find the leaking issue of the code below?
If we visit the port 3000 again and again the service will leak memory.

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    _ = http.ListenAndServe(":3000", http.HandlerFunc(handle))
}

func handle(_ http.ResponseWriter, _ *http.Request) {
    c := make(chan int)
    go produce(c)
    go consume(c)
}

func produce(c chan int) {
    for i := range "..." {
        c <- i
    }
}

func consume(c chan int) {
    for i := range c {
        fmt.Println(i)
    }
}
```

We can use gotrace to locate the leak point, let's create a `main_test.go` file:

```go
package main

import (
    "testing"

    "github.com/ysmood/gotrace"
)

func TestHandle(t *testing.T) {
    // Just add one line before the standard test case
    gotrace.CheckLeak(t, 0)

    handle(nil, nil)
}
```

Run `GODEBUG="tracebackancestors=10" go test`, wait for 3 seconds you will see the below output:

```txt
--- FAIL: TestHandle (3.00s)
    main_test.go:10: leaking goroutines: goroutine 6 [chan receive]:
        play.consume(0x0)
            /Users/ys/repos/play/default/main.go:25 +0x74
```

The `range` on line 25 get stuck, the reason is `chan receive`, it keeps waiting for new messages from `c`, but the `produce` function has already exited, it will not send messages to `c` anymore.

We have several ways to fix it, such as change the `produce` to:

```go
func produce(c chan int) {
    for i := range "..." {
        c <- i
    }
    close(c)
}
```

Run `go test` again, the test should exit immediately without errors.
