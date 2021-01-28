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
