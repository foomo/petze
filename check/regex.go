package check

import (
	"github.com/dreadl0ck/petze/config"
	"regexp"
	"strings"
)

func Regex(data []byte, selector string, expect config.Expect) (ok bool, info string) {

	regex, errCompile := regexp.Compile(selector)
	if errCompile != nil {
		return false, "could not compile regex '" + selector + "'"
	}

	res := regex.FindAll(data, -1)
	switch true {
	case expect.Min != nil || expect.Max != nil || expect.Count != nil:
		return checkMinMaxCount(expect, int64(len(res)))
	case expect.Equals != nil:
		for _, res := range res {
			if string(res) == expect.Equals {
				return true, "regex match found"
			}
		}
		return false, "could not find regex result equals"
	case expect.Contains != "":
		for _, res := range res {
			if strings.Contains(string(res), expect.Contains) {
				return true, "regex contains substring found"
			}
		}
		return false, "could not find regex result contains"
	default:
		return false, "comparator not implemented for regex"
	}
}
