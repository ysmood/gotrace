package gotrace

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func defaultMaxWait(t time.Duration) time.Duration {
	if t <= 0 {
		return 3 * time.Second
	}
	return t
}

// Check if there's goroutine leak
func Check(maxWait time.Duration, ignores ...Ignore) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultMaxWait(maxWait))
	defer cancel()

	if traces := Wait(ctx, ignores...); traces.Any() {
		return fmt.Errorf("leaking goroutines: %s", traces)
	}
	return nil
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

// CheckMain reports error if goroutines are leaking after all tests are done. Default timeout is 3s.
// It's powerful but less accurate than Check, if you only use CheckMain it will be hard to tell which test
// is the cause of the leak.
func CheckMain(m M, maxWait time.Duration, ignores ...Ignore) {
	code := m.Run()
	if code != 0 {
		os.Exit(code)
	}

	if err := Check(maxWait, ignores...); err != nil {
		log.Fatal(err)
	}
}

// CheckTest reports error if the test is leaking goroutine.
// Default timeout is 3s. Default ignore is gotrace.IgnoreCurrent() .
//
// This check will become meaningless if t.Parallel() is called for multiple tests,
// because the test framework will execute tests at the same time which makes it impossible to
// write a correct ignore function to detect which goroutine is spawned by current test.
// But you can still use CheckMain to check leak, because it runs after all tests are settled.
func CheckTest(t T, maxWait time.Duration, ignores ...Ignore) {
	t.Helper()

	if len(ignores) == 0 {
		ignores = []Ignore{IgnoreCurrent()}
	}

	t.Cleanup(func() {
		t.Helper()

		if t.Failed() {
			return
		}

		if err := Check(maxWait, ignores...); err != nil {
			t.Logf("%v", err)
		}
	})
}
