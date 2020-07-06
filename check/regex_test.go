package check

import (
	"github.com/dreadl0ck/petze/config"
	"testing"
)

var regexTestCases = []struct {
	expect  config.Expect
	data    string
	regex   string
	ok      bool
	info    string
	message string
}{
	{
		data:   "<message>some fake body which body needs body three times</message>",
		regex:  `body`,
		expect: expectMin(3), ok: true, info: "", message: "check minimum matches on regex",
	},
	{
		data:   "<message>some fake body which body needs body three times</message>",
		regex:  `body`,
		expect: expectMax(3), ok: true, info: "", message: "check max number of occurrences",
	},
	{
		data:   "<message>some fake body which body needs body three times</message>",
		regex:  `body`,
		expect: expectCount(3), ok: true, info: "", message: "check exact match count",
	},
	{
		data:   "<message>some 123456 body which body needs body three times</message>",
		regex:  `\d+`,
		expect: expectEquals("123456"), ok: true, info: "", message: "check find exact value",
	},
	{
		data:   "<message>some 123456 body which body needs body three times</message>",
		regex:  `\d+`,
		expect: expectContains("234"), ok: true, info: "", message: "check find exact value",
	},
	{
		data:   "invalid regex",
		regex:  `[`,
		expect: expectMin(1), ok: false, info: "could not compile regex '['", message: "invalid regex check",
	},
	{
		data:   "missing regex equals",
		regex:  `derp`,
		expect: expectEquals("derp"), ok: false, info: "could not find regex result equals", message: "missing equals",
	},
	{
		data:   "missing regex contains",
		regex:  `herp`,
		expect: expectContains("herp"), ok: false, info: "could not find regex result contains", message: "missing contains",
	},
	{
		data:   "unknown comparator",
		regex:  `herp`,
		expect: config.Expect{}, ok: false, info: "comparator not implemented for regex", message: "missing contains",
	},
}

func TestCheckRegex(t *testing.T) {
	for _, test := range regexTestCases {
		ok, info := Regex([]byte(test.data), test.regex, test.expect)
		if ok != test.ok && info != test.info {
			t.Error(test.message)
		}
	}
}

func expectEquals(value string) config.Expect {
	return config.Expect{Equals: value}
}

func expectContains(value string) config.Expect {
	return config.Expect{Contains: value}
}

func expectMin(min int) config.Expect {
	return config.Expect{Min: &[]int64{int64(min)}[0]}
}

func expectMax(max int) config.Expect {
	return config.Expect{Max: &[]int64{int64(max)}[0]}
}

func expectCount(count int) config.Expect {
	return config.Expect{Max: &[]int64{int64(count)}[0]}
}
