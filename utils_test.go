package gotrace_test

import (
	"context"
	"testing"
	"time"

	"github.com/ysmood/gotrace"
)

func TestMergeStackInfo(t *testing.T) {
	wait := make(chan int)

	fn := func(int) {
		<-wait
	}

	ig := gotrace.IgnoreCurrent()

	for i := range "..." {
		go fn(i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	info := gotrace.Wait(ctx, ig).String()

	if info[:3] != "[3]" {
		t.Fail()
	}

	close(wait)
}

func TestGetElided(t *testing.T) {
	old := gotrace.GetStack
	defer func() {
		gotrace.GetStack = old
	}()

	gotrace.GetStack = func(all bool) string {
		return `goroutine 117 [runnable]:
reflect.resolveTypeOff(0x7cd5a0, 0x45f40, 0x7bef40)
	D:/Go16/src/runtime/runtime1.go:504 +0x3a
reflect.(*rtype).typeOff(...)
	D:/Go16/src/reflect/type.go:690
reflect.(*rtype).ptrTo(0x7cd5a0, 0x6)
	D:/Go16/src/reflect/type.go:1384 +0x36c
reflect.Value.Addr(0x7cd5a0, 0xc0001f6380, 0x198, 0xa489e0, 0x7cd5a0, 0x198)
	D:/Go16/src/reflect/value.go:276 +0x3d
encoding/json.(*decodeState).array(0xc0000a42c0, 0x7c1360, 0xc00013a1f8, 0x197, 0xc0000a42e8, 0x5b)
	D:/Go16/src/encoding/json/decode.go:558 +0x1b4
...additional frames elided...
created by main.testFn.func2
	F:/test/fold/go-rod/debugg.go:518 +0x65e`
	}

	if gotrace.Get(true)[0].Stacks[5].Func != "main.testFn.func2" {
		t.Fail()
	}
}

func TestIgnoreNonChildrenPanic(t *testing.T) {
	old := gotrace.TraceAncestorsEnabled
	gotrace.TraceAncestorsEnabled = false
	t.Cleanup(func() {
		gotrace.TraceAncestorsEnabled = old
	})

	defer func() {
		r := recover()
		if r.(string) != `You must set GODEBUG="tracebackancestors=N", N should be a big enough integer, such as 1000` {
			t.Fail()
		}
	}()

	gotrace.IgnoreNonChildren()
}
