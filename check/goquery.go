package check

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/foomo/petze/config"
)

func Goquery(doc *goquery.Document, selector string, expect config.Expect) (ok bool, info string) {
	selection := doc.Find(selector)
	length := selection.Length()
	switch true {
	case expect.Min != nil:
		ok = length >= int(*expect.Min)
		if !ok {
			info = fmt.Sprint("min:", length, "<", *expect.Min)
		}
		return
	case expect.Max != nil:
		ok = length <= int(*expect.Max)
		if !ok {
			info = fmt.Sprint("max", length, ">", *expect.Max)
		}
		return
	case expect.Count != nil:
		ok = length == int(*expect.Count)
		if !ok {
			info = fmt.Sprint("count", length, "!=", *expect.Count)
		}
		return
	case expect.Contains != "":
		info = "contains is not implemented"
	case expect.Equals != nil:
		info = "equals is not implemented"
	}
	return
}
