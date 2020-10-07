package gotrace

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Stack info
type Stack struct {
	Func string
	Loc  string
}

// Trace of one goroutine
type Trace struct {
	Raw         string
	GoroutineID int64
	WaitReason  string // https://github.com/golang/go/blob/874b3132a84cf76da6a48978826c04c380a37a50/src/runtime/runtime2.go#L997
	Stacks      []Stack
}

// String interface for fmt
func (t Trace) String() string {
	return t.Raw
}

// Traces of goroutines
type Traces []Trace

// Any item exists in the list
func (list Traces) Any() bool {
	return len(list) > 0
}

// String interface for fmt
func (list Traces) String() string {
	out := ""
	for _, t := range list {
		out += t.Raw + "\n\n"
	}
	return out
}

var regGoroutine = regexp.MustCompile(`^goroutine (\d+) \[(.+)\]:`)
var regFunc = regexp.MustCompile(`^(.+?)(\([\w, ]+\))?$`)
var regLoc = regexp.MustCompile(`^\t(.+)( \+0x\w+)?$`)

// Get the Trace of the calling goroutine.
// If all is true, all other goroutines' Traces will be appended into the result too.
func Get(all bool) Traces {
	rawList := strings.Split(getStack(all), "\n\n")
	list := []Trace{}

	for _, raw := range rawList {
		lines := strings.Split(raw, "\n")

		t := Trace{
			Raw:    raw,
			Stacks: []Stack{},
		}

		for j, line := range lines {
			if j == 0 {
				ms := regGoroutine.FindStringSubmatch(line)
				id, _ := strconv.ParseInt(ms[1], 10, 64)
				t.GoroutineID = id
				t.WaitReason = ms[2]
			}

			if j%2 == 1 {
				t.Stacks = append(t.Stacks, Stack{
					regFunc.FindStringSubmatch(line)[1],
					regLoc.FindStringSubmatch(lines[j+1])[1],
				})
			}
		}

		list = append(list, t)
	}

	return list
}

func getStack(all bool) string {
	for i := 1024 * 1024; ; i *= 2 {
		buf := make([]byte, i)
		if n := runtime.Stack(buf, all); n < i {
			return string(buf[:n-1])
		}
	}
}
