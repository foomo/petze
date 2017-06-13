package check

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/foomo/petze/config"
)

func Goquery(doc *goquery.Document, selector string, expect config.Expect) (ok bool, info string) {
	switch true {
	case expect.Max != nil, expect.Min != nil, expect.Count != nil:
		return checkMinMaxCount(expect, int64(doc.Find(selector).Length()))
	case expect.Contains != "":
		info = "contains is not implemented"
	case expect.Equals != nil:
		info = "equals is not implemented"
	}
	return
}
