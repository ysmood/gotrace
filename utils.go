package gotrace

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"time"
)

// GetStack of current runtime
var GetStack = func(all bool) string {
	for i := 1024 * 1024; ; i *= 2 {
		buf := make([]byte, i)
		if n := runtime.Stack(buf, all); n < i {
			return string(buf[:n-1])
		}
	}
}

// Ignore returns true to ignore t
type Ignore func(t *Trace) bool

// IgnoreCurrent running goroutines
func IgnoreCurrent() Ignore {
	return IgnoreList(Get(true))
}

// TraceAncestorsEnabled returns true if GODEBUG="tracebackancestors=N" is set
var TraceAncestorsEnabled = regexp.MustCompile(`tracebackancestors=\d+`).MatchString(os.Getenv("GODEBUG"))

// IgnoreNonChildren goroutines
func IgnoreNonChildren() Ignore {
	if !TraceAncestorsEnabled {
		panic(`You must set GODEBUG="tracebackancestors=N", N should be a big enough integer, such as 1000`)
	}

	id := Get(false)[0].GoroutineID
	return func(t *Trace) bool {
		return !t.HasParent(id)
	}
}

// IgnoreList of traces
func IgnoreList(list Traces) Ignore {
	return func(t *Trace) bool {
		for _, item := range list {
			if t.GoroutineID == item.GoroutineID {
				return true
			}
		}
		return false
	}
}

// IgnoreFuncs ignores a Trace if it's first Stack's Func equals one of the names.
func IgnoreFuncs(names ...string) Ignore {
	return func(t *Trace) bool {
		for _, name := range names {
			if t.Stacks[0].Func == name {
				return true
			}
		}
		return false
	}
}

// CombineIgnores into one
func CombineIgnores(list ...Ignore) Ignore {
	return func(t *Trace) bool {
		for _, i := range list {
			if i(t) {
				return true
			}
		}
		return false
	}
}

// Backoff is the default algorithm for sleep backoff
var Backoff = func(t time.Duration) time.Duration {
	const maxSleep = 300 * time.Millisecond
	if t == 0 {
		return time.Microsecond
	}

	t *= 2

	if t > maxSleep {
		return maxSleep
	}

	return t
}

// Wait uses Backoff for WaitWithBackoff
func Wait(ctx context.Context, ignores ...Ignore) (remain Traces) {
	return WaitWithBackoff(ctx, Backoff, ignores...)
}

// WaitWithBackoff algorithm. Wait for other goroutines that are not ignored to exit. It returns the ones that are still active.
// It keeps counting the active goroutines that are not ignored, if the number is zero return.
func WaitWithBackoff(ctx context.Context, backoff func(time.Duration) time.Duration, ignores ...Ignore) (remain Traces) {
	sleep := backoff(0)
	ignore := CombineIgnores(ignores...)

	for {
		remain = Get(true)[1:].Filter(ignore)

		if len(remain) == 0 {
			return
		}

		sleep = backoff(sleep)
		tmr := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			tmr.Stop()
			return
		case <-tmr.C:
		}
	}
}

var nullCancel = func() {}

// Timeout shortcut for context.WithTimeout(context.Background(), d)
func Timeout(d time.Duration) context.Context {
	ctx, c := context.WithTimeout(context.Background(), d)
	nullCancel()
	nullCancel = c
	return ctx
}

// Signal to cancel the returned context, default signal is CTRL+C .
func Signal(signals ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	if len(signals) == 0 {
		signals = append(signals, os.Interrupt)
	}

	signal.Notify(c, signals...)
	go func() {
		<-c
		signal.Stop(c)
		cancel()
	}()

	return ctx
}

// Traces of goroutines
type Traces []*Trace

// Any item exists in the list
func (list Traces) Any() bool {
	return len(list) > 0
}

// Filter returns the remain Traces that are ignored
func (list Traces) Filter(ignore Ignore) Traces {
	remain := Traces{}
	for _, s := range list {
		if ignore(s) {
			continue
		}
		remain = append(remain, s)
	}
	return remain
}

// String interface for fmt. It will merge similar trace together and print counts.
func (list Traces) String() string {
	type group struct {
		count int
		trace *Trace
	}

	// group by stack
	groups := map[string]*group{}
	for _, t := range list {
		if g, has := groups[t.typeKey]; has {
			g.count++
		} else {
			groups[t.typeKey] = &group{count: 1, trace: t}
		}
	}

	out := ""
	for _, g := range groups {
		if g.count > 1 {
			out += fmt.Sprintf("[%d] ", g.count) + g.trace.Raw + "\n\n"
		} else {
			out += g.trace.Raw + "\n\n"
		}
	}
	return out
}
