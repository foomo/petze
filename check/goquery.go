package check

import (
	"reflect"

	"github.com/PuerkitoBio/goquery"
	"github.com/dreadl0ck/petze/config"
)

func Goquery(doc *goquery.Document, selector string, expect config.Expect) (ok bool, info string) {
	switch true {
	case expect.Max != nil, expect.Min != nil, expect.Count != nil:
		return checkMinMaxCount(expect, int64(doc.Find(selector).Length()))
	case expect.Contains != "":
		info = "contains is not implemented"
	case expect.Equals != nil:
		info = "equals is not implemented"
		expectRefl := reflect.ValueOf(expect.Equals)
		switch expectRefl.Kind().String() {
		case "string":
			//res := doc.Find(selector)
			//fmt.Println(selector, "length", res.Length(), "text", res.Text())
			actualString := doc.Find(selector).Text()
			expectString := expect.Equals.(string)
			return checkExpectStringEquals(expect, expectString, actualString)
		default:
			info += " for kind " + expectRefl.Kind().String()
		}
	}
	return
}
