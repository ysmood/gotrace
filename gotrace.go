package gotrace

import (
	"crypto/md5"
	"fmt"
	"regexp"
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
	Raw                  string
	GoroutineID          int64
	GoroutineAncestorIDs []int64 // Need GODEBUG="tracebackancestors=N" to be set
	WaitReason           string  // https://github.com/golang/go/blob/874b3132a84cf76da6a48978826c04c380a37a50/src/runtime/runtime2.go#L997
	Stacks               []Stack

	typeKey string // hash sum of WaitReason and Stacks
}

// String interface for fmt
func (t Trace) String() string {
	return t.Raw
}

// HasParent goroutine id
func (t Trace) HasParent(id int64) bool {
	for _, pid := range t.GoroutineAncestorIDs {
		if pid == id {
			return true
		}
	}
	return false
}

var regGoroutine = regexp.MustCompile(`^goroutine (\d+) \[(.+)\]:`)
var regParent = regexp.MustCompile(`^\[originating from goroutine (\d+)\]:`)
var regFunc = regexp.MustCompile(`^(?:created by )?([^()]+)`)
var regLoc = regexp.MustCompile(`^\t(.*)( \+0x\w+)?$`)

// Get the Trace of the calling goroutine.
// If all is true, all other goroutines' Traces will be appended into the result too.
func Get(all bool) Traces {
	rawList := strings.Split(GetStack(all), "\n\n")
	list := []*Trace{}

	for _, raw := range rawList {
		lines := strings.Split(raw, "\n")

		t := &Trace{
			Raw:                  raw,
			Stacks:               []Stack{},
			GoroutineAncestorIDs: []int64{},
		}

		typeKey := md5.New()

		l := len(lines) - 3
		if l > -1 && lines[l] == "...additional frames elided..." {
			lines = append(lines[:l], lines[l+1:]...)
		}

		ancestor := false

		for i := 0; i < len(lines); i++ {
			l := lines[i]

			if i == 0 {
				ms := regGoroutine.FindStringSubmatch(l)
				id, _ := strconv.ParseInt(ms[1], 10, 64)
				t.GoroutineID = id
				t.WaitReason = ms[2]

				_, _ = typeKey.Write([]byte(t.WaitReason))
				continue
			} else if i != 0 && l[len(l)-1] == ':' {
				ancestor = true
				ms := regParent.FindStringSubmatch(l)
				id, _ := strconv.ParseInt(ms[1], 10, 64)
				t.GoroutineAncestorIDs = append(t.GoroutineAncestorIDs, id)
			}

			if ancestor {
				continue
			}

			s := Stack{
				regFunc.FindStringSubmatch(l)[1],
				regLoc.FindStringSubmatch(lines[i+1])[1],
			}
			t.Stacks = append(t.Stacks, s)

			_, _ = typeKey.Write([]byte(s.Func))
			_, _ = typeKey.Write([]byte(s.Loc))

			i++
		}

		t.typeKey = fmt.Sprintf("%x", typeKey.Sum(nil))

		list = append(list, t)
	}

	return list
}
