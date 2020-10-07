package gotrace

import (
	"context"
	"os"
	"os/signal"
	"time"
)

// Ignore filter
type Ignore func(Trace) bool

// IgnoreCurrent Trace list
func IgnoreCurrent() Ignore {
	return IgnoreList(Get(true))
}

// IgnoreList of Trace
func IgnoreList(list Traces) Ignore {
	return func(t Trace) bool {
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
	return func(t Trace) bool {
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
	return func(t Trace) bool {
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
		remain = []Trace{}
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

var nullCancel func()

// Timeout shortcut for context.WithTimeout(context.Background(), d)
func Timeout(d time.Duration) context.Context {
	ctx, c := context.WithTimeout(context.Background(), d)
	nullCancel = c
	return ctx
}

// Signal to cancel the returned context, default signal is CTRL+C .
func Signal(signals ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal)
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
