package testleak

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ysmood/gotrace"
)

const leakErrMsg = "leaking goroutines:"

func defaultMaxWait(t time.Duration) time.Duration {
	if t <= 0 {
		return 3 * time.Second
	}
	return t
}

// CheckMain reports error if goroutines are leaking after all tests are done. Default timeout is 3s.
// It's powerful but less accurate than Check, if you only use CheckMain it will be hard to tell which test
// is the cause of the leak.
func CheckMain(m *testing.M, maxWait time.Duration, ignores ...gotrace.Ignore) {
	code := m.Run()
	if code != 0 {
		os.Exit(code)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMaxWait(maxWait))
	defer cancel()

	if traces := gotrace.Wait(ctx, ignores...); traces.Any() {
		log.Fatalln(leakErrMsg, traces)
	}
}

// Check reports error if the test is leaking goroutine.
// Default timeout is 3s. Default ignore is gotrace.IgnoreCurrent() .
//
// This check will become useless if t.Parallel() is called for multiple tests,
// because the test framework will execute tests at the same time which makes it impossible to
// write a correct ignore function to detect which goroutine is spawned by current test.
// But you can still use CheckMain to check leak, because it runs after all tests are settled.
func Check(t *testing.T, maxWait time.Duration, ignores ...gotrace.Ignore) {
	t.Helper()

	if len(ignores) == 0 {
		ignores = []gotrace.Ignore{gotrace.IgnoreCurrent()}
	}

	t.Cleanup(func() {
		t.Helper()

		if t.Failed() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultMaxWait(maxWait))
		defer cancel()

		if traces := gotrace.Wait(ctx, ignores...); traces.Any() {
			t.Error(leakErrMsg, traces)
		}
	})
}
