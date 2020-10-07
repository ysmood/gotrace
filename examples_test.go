package gotrace_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ysmood/gotrace"
)

func ExampleGet() {
	list := gotrace.Get(true)

	fmt.Println("goroutine count:", len(list))
	fmt.Println("id of current:", list[0].GoroutineID)
	fmt.Println("caller of current:", list[0].Stacks[2].Func)

	// Output:
	//
	// goroutine count: 2
	// id of current: 1
	// caller of current: github.com/ysmood/gotrace_test.ExampleGet()
}

func ExampleIgnoreCurrent() {
	ignore := gotrace.IgnoreCurrent()

	go func() {
		time.Sleep(time.Second)
	}()

	start := time.Now()
	gotrace.Wait(context.TODO(), ignore)
	end := time.Since(start)

	if end > time.Second {
		fmt.Println("waited for 1 second")
	}

	// Output: waited for 1 second
}

func ExampleIgnoreFuncs() {
	ignore := gotrace.IgnoreFuncs("internal/poll.runtime_pollWait")

	go func() {
		time.Sleep(time.Second)
	}()

	start := time.Now()
	gotrace.Wait(context.TODO(), ignore)
	end := time.Since(start)

	if end > time.Second {
		fmt.Println("waited for 1 second")
	}

	// Output: waited for 1 second
}

func ExampleCombineIgnores() {
	ignore := gotrace.CombineIgnores(
		gotrace.IgnoreCurrent(),
		func(t gotrace.Trace) bool {
			return strings.Contains(t.Raw, "ExampleCombineIgnores.func2")
		},
	)

	go func() {
		time.Sleep(2 * time.Second)
	}()

	go func() {
		time.Sleep(time.Second)
	}()

	start := time.Now()
	gotrace.Wait(context.TODO(), ignore)
	end := time.Since(start)

	if time.Second < end && end < 2*time.Second {
		fmt.Println("only waits for the second goroutine")
	}

	// Output: only waits for the second goroutine
}

func ExampleTraces_String() {
	go func() {
		time.Sleep(time.Second)
	}()

	traces := gotrace.Wait(gotrace.Timeout(0))

	str := fmt.Sprintf("%v %v", traces[0], traces)

	fmt.Println(strings.Contains(str, "gotrace_test.ExampleTraces_String"))

	// Output: true
}

func ExampleSignal() {
	go func() {
		traces := gotrace.Wait(gotrace.Signal())
		fmt.Println(strings.Contains(traces.String(), "gotrace_test.ExampleSignal"))
	}()

	time.Sleep(100 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)

	time.Sleep(100 * time.Millisecond)

	// Output: true
}
