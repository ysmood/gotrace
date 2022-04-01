package gotrace

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// Exit to os.Exit
var Exit = os.Exit

// CheckWithContext if there's goroutine leak
func CheckWithContext(ctx context.Context, ignores ...Ignore) error {
	if traces := Wait(ctx, ignores...); traces.Any() {
		return fmt.Errorf("leaking goroutines: %s", traces)
	}
	return nil
}

// Check if there's goroutine leak
func Check(timeout time.Duration, ignores ...Ignore) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout(timeout))
	defer cancel()

	return CheckWithContext(ctx, ignores...)
}

// M interface for testing.M
type M interface {
	Run() int
}

// T interface for testing.T
type T interface {
	Helper()
	Fail()
	Failed() bool
	Cleanup(f func())
	Logf(format string, args ...interface{})
}

// CheckMainLeak reports error if goroutines are leaking after all tests are done. Default timeout is 3s.
// It's powerful but less accurate than Check, if you only use CheckMainLeak it will be hard to tell which test
// is the cause of the leak.
func CheckMainLeak(m M, timeout time.Duration, ignores ...Ignore) {
	code := m.Run()
	if code != 0 {
		Exit(code)
		return
	}

	if err := Check(timeout, ignores...); err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("%s:%d %v\n", file, line, err)
		Exit(1)
	}
}

// CheckLeak reports error if the test is leaking goroutine.
// Default timeout is 3s. Default ignore is gotrace.IgnoreNonChildren() .
func CheckLeak(t T, timeout time.Duration, ignores ...Ignore) {
	t.Helper()

	if len(ignores) == 0 {
		ignores = []Ignore{IgnoreNonChildren()}
	}

	t.Cleanup(func() {
		t.Helper()

		if t.Failed() {
			return
		}

		if err := Check(timeout, ignores...); err != nil {
			t.Logf("%v", err)
			t.Fail()
		}
	})
}

func defaultTimeout(t time.Duration) time.Duration {
	if t <= 0 {
		return 3 * time.Second
	}
	return t
}
