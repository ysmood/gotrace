package gotrace

import (
	"crypto/md5"
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

	typeKey string // hash sum of WaitReason and Stacks
}

// String interface for fmt
func (t Trace) String() string {
	return t.Raw
}

var regGoroutine = regexp.MustCompile(`^goroutine (\d+) \[(.+)\]:`)
var regFunc = regexp.MustCompile(`^(.+?)(\([\w, ]+\))?$`)
var regLoc = regexp.MustCompile(`^\s+([^\s].*)( \+0x\w+)?$`)

// Get the Trace of the calling goroutine.
// If all is true, all other goroutines' Traces will be appended into the result too.
func Get(all bool) Traces {
	rawList := strings.Split(getStack(all), "\n\n")
	list := []*Trace{}

	for _, raw := range rawList {
		lines := strings.Split(raw, "\n")

		t := &Trace{
			Raw:    raw,
			Stacks: []Stack{},
		}

		typeKey := md5.New()

		for j, line := range lines {
			if j == 0 {
				ms := regGoroutine.FindStringSubmatch(line)
				id, _ := strconv.ParseInt(ms[1], 10, 64)
				t.GoroutineID = id
				t.WaitReason = ms[2]

				_, _ = typeKey.Write([]byte(t.WaitReason))
			}

			if j%2 == 1 {
				s := Stack{
					regFunc.FindStringSubmatch(line)[1],
					regLoc.FindStringSubmatch(lines[j+1])[1],
				}
				t.Stacks = append(t.Stacks, s)

				_, _ = typeKey.Write([]byte(s.Func))
				_, _ = typeKey.Write([]byte(s.Loc))
			}
		}

		t.typeKey = string(typeKey.Sum(nil))

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
