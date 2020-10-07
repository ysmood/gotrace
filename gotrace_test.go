package gotrace_test

import (
	"testing"

	"github.com/ysmood/gotrace/pkg/testleak"
)

func TestMain(m *testing.M) {
	testleak.CheckMain(m, 0) // make sure no leak after all tests
}

func TestCase(t *testing.T) {
	testleak.Check(t, 0) // just add one line before each test
}
