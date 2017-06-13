package check

import (
	"fmt"

	"github.com/foomo/petze/config"
)

func checkMinMaxCount(expect config.Expect, length int64) (ok bool, info string) {
	switch true {
	case expect.Min != nil:
		ok = length >= *expect.Min
		if !ok {
			info = fmt.Sprint("min actual: ", length, " < expected: ", *expect.Min)
		}
		return
	case expect.Max != nil:
		ok = length <= *expect.Max
		if !ok {
			info = fmt.Sprint("max actual: ", length, " > expected: ", *expect.Max)
		}
		return
	case expect.Count != nil:
		ok = length == *expect.Count
		if !ok {
			info = fmt.Sprint("count actual: ", length, " != expected: ", *expect.Count)
		}
		return
	default:
		panic("this is a programming error - check your usage of minMaxCount")
	}
}
