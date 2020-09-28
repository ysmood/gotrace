package gotrace_test

import (
	"testing"

	"github.com/ysmood/gotrace/pkg/testleak"
)

func TestMain(m *testing.M) {
	testleak.CheckMain(m, 0) // make sure not leak after all tests
}

func TestCase(t *testing.T) {
	// Just add one line before each test.
	// If you feel it's boring to add this line to every test, you can use the TestSuite style.
	testleak.Check(t, 0)
}
