package gotrace

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

// Ignore filter
type Ignore func(*Trace) bool

// IgnoreCurrent Trace list
func IgnoreCurrent() Ignore {
	return IgnoreList(Get(true))
}

// IgnoreList of Trace
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

// Wait for other goroutines that are not ignored to exit. It returns the ones that are still active.
// It keeps counting the active goroutines that are not ignored, if the number is zero return.
func Wait(ctx context.Context, ignores ...Ignore) (remain Traces) {
	sleep := time.Microsecond
	const maxSleep = 300 * time.Millisecond
	ignore := CombineIgnores(ignores...)

	for {
		remain = []*Trace{}
		list := Get(true)[1:]
		for _, s := range list {
			if ignore(s) {
				continue
			}
			remain = append(remain, s)
		}

		if len(remain) == 0 {
			return
		}

		// backoff
		sleep *= 2
		if sleep > maxSleep {
			sleep = maxSleep
		}

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
