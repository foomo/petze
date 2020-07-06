package check

import (
	"reflect"

	"github.com/Jeffail/gabs"
	"github.com/JumboInteractiveLimited/jsonpath"
	"github.com/dreadl0ck/petze/config"

	log "github.com/sirupsen/logrus"
)

func JSONPath(jsonBytes []byte, selector string, expect config.Expect) (ok bool, info string) {
	info = "check not implemented"

	paths, errParsePaths := jsonpath.ParsePaths(selector)
	if errParsePaths != nil {
		info = "could not parse json paths : " + errParsePaths.Error()
		return
	}

	eval, errEval := jsonpath.EvalPathsInBytes(jsonBytes, paths)
	if errEval != nil {
		info = "error in json path : " + errEval.Error()
		return
	}

	result, evalOK := eval.Next()
	if !evalOK {
		info = "could not eval jsonpath: " + selector
		return
	}

	if len(result.Value) == 0 {
		info = "no result for " + selector
		return
	}

	json, jsonErr := gabs.ParseJSON(result.Value)
	if jsonErr != nil {
		info = "could not parse json: " + jsonErr.Error() + " " + string(result.Value)
		return
	}

	//fmt.Println(json.Data()) // true -> show keys in pretty string
	data := json.Data()
	refl := reflect.ValueOf(data)
	length := int64(-1)
	resultIsString := false
	resultString := ""
	switch refl.Kind().String() {
	case "string":
		resultIsString = true
		resultString = data.(string)
	case "slice":
		length = int64(len(data.([]interface{})))
	default:
		log.Warn("wtf would that be", refl.Type().String(), refl.Kind().String())
	}

	if eval.Error != nil {
		info = "could not evaluate json path " + eval.Error.Error()
		return
	}

	switch true {
	case expect.Min != nil, expect.Max != nil, expect.Count != nil:
		return checkMinMaxCount(expect, int64(length))
	case expect.Equals != nil:
		expected := ""
		if reflect.ValueOf(expect.Equals).Type().String() != "string" {
			info = "jsonpath can only compare to string"
			return
		}
		expected = expect.Equals.(string)
		if !resultIsString {
			info = "result is not a string"
			return
		}
		if resultString == expected {
			ok = true
			return
		}
		info = "actual: " + resultString + " != expected: " + expected
		return
	}
	return
}
