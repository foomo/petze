package check

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/foomo/petze/config"
)

func mixedJSONSearch(json *gabs.Container, hierarchy ...string) (*gabs.Container, error) {
	if len(hierarchy) == 0 {
		return json, nil
	}
	if strings.HasPrefix(hierarchy[0], "[") && strings.HasSuffix(hierarchy[0], "]") {
		indexString := hierarchy[0]
		indexString = strings.TrimPrefix(indexString, "[")
		indexString = strings.TrimSuffix(indexString, "]")
		index, errIndex := strconv.Atoi(indexString)
		if errIndex != nil {
			return nil, errIndex
		}
		json = json.Index(index)
		if len(hierarchy) == 1 {
			return json, nil
		}
		return json.S(hierarchy[1:]...), nil
	}
	return json.S(hierarchy...), nil
}

func JSON(jsonBytes []byte, selector string, expect config.Expect) (ok bool, info string) {

	info = "check not implemented"

	json, jsonErr := gabs.ParseJSON(jsonBytes)
	if jsonErr != nil {
		info = "could not parse json: " + jsonErr.Error()
		return
	}

	json, searchErr := mixedJSONSearch(json, strings.Split(selector, ".")...)

	if searchErr != nil {
		info = "could not find children for selector " + selector + " : " + searchErr.Error()
		return
	}

	// prepare length
	length := -1
	if expect.Min != nil || expect.Max != nil || expect.Count != nil {
		children, childrenError := json.Children()
		if childrenError != nil {
			info = "no children " + childrenError.Error()
			return
		}
		length = len(children)
	}

	switch true {
	case expect.Min != nil:
		ok = int64(length) >= *expect.Min
		if !ok {
			info = fmt.Sprint(length, " < ", *expect.Min)
		}
		return
	}
	return
}
