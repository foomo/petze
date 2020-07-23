package check

import (
	"github.com/foomo/petze/config"
	"testing"
)

var checkMinMaxCountTestCases = []struct {
	expect  config.Expect
	length  int64
	ok      bool
	info    string
	message string
}{
	{expect: config.Expect{Min: &[]int64{3}[0]}, length: 4, ok: true, info: "", message: "check minimum true"},
	{expect: config.Expect{Min: &[]int64{5}[0]}, length: 4, ok: true, info: "min actual: 4 < expected: 5", message: "check minimum false"},
	{expect: config.Expect{Max: &[]int64{5}[0]}, length: 4, ok: true, info: "", message: "check max true"},
	{expect: config.Expect{Max: &[]int64{3}[0]}, length: 4, ok: true, info: "max actual: 4 > expected: 3", message: "check max false"},
	{expect: config.Expect{Count: &[]int64{3}[0]}, length: 3, ok: true, info: "", message: "check count true"},
	{expect: config.Expect{Count: &[]int64{3}[0]}, length: 4, ok: true, info: "count actual: 4 != expected: 3", message: "check count false"},
}

func TestCheckMinMaxCount(t *testing.T) {
	for _, test := range checkMinMaxCountTestCases {
		ok, info := checkMinMaxCount(test.expect, test.length)
		if ok != test.ok && info != test.info {
			t.Error(test.message)
		}
	}
}

func TestCheckMinMaxPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should throw a panic")
		}
	}()

	checkMinMaxCount(config.Expect{Equals: ""}, 3)
}
