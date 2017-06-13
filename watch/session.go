package watch

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"time"

	"io/ioutil"

	"io"

	"encoding/json"

	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/foomo/petze/check"
	"github.com/foomo/petze/config"
)

func runSession(service *config.Service, r *Result, client *http.Client) error {
	//log.Println("running session with session length:", len(service.Session))
	// utils.JSONDump(service)
	endPointURL, errURL := service.GetURL()
	if errURL != nil {
		return errors.New("can not run session: " + errURL.Error())
	}
	for _, call := range service.Session {
		// copy URL
		callURL := &url.URL{}
		*callURL = *endPointURL

		uriURL, errURIURL := call.GetURL()
		if errURIURL != nil {
			return errURIURL
		}

		callURL.Path = uriURL.Path
		callURL.RawQuery = uriURL.RawQuery

		var body io.Reader
		method := http.MethodGet
		if call.Method != "" {
			method = call.Method
		}
		if call.Data != nil {
			dataBytes, errDataBytes := json.Marshal(call.Data)
			if errDataBytes != nil {
				return errors.New("could not encode data bytes: " + errDataBytes.Error())
			}
			body = bytes.NewBuffer(dataBytes)
		}

		req, errNewRequest := http.NewRequest(method, callURL.String(), body)
		if errNewRequest != nil {
			return errNewRequest
		}
		start := time.Now()
		response, errResponse := client.Do(req)
		if errResponse != nil {
			return errResponse
		}
		duration := time.Since(start)

		for _, check := range call.Check {
			errCheck := checkResponse(r, call, response, check, duration)
			if errCheck != nil {
				return errCheck
			}
		}

	}
	return nil
}

func checkResponse(r *Result, call config.Call, response *http.Response, chk config.Check, callDuration time.Duration) error {
	//	log.Println("checking response", chk)

	addError := func(err string, t ErrorType) {
		r.addError(errors.New(err), t, chk.Comment)
	}
	switch true {
	case chk.Duration > 0:
		if callDuration > chk.Duration {
			addError(fmt.Sprint("call duration ", callDuration, " exceeded ", chk.Duration), ErrorTypeServerTooSlow)
		}
	case chk.Goquery != nil:
		// go query
		doc, errDoc := goquery.NewDocumentFromResponse(response)
		if errDoc != nil {
			return errDoc
		}
		for selector, expect := range chk.Goquery {
			ok, info := check.Goquery(doc, selector, expect)
			if !ok {
				//log.Println("no match", selector, expect)
				addError(info, ErrorTypeGoQueryMismatch)
			}
		}
	case chk.Data != nil:
		contentType := config.ContentTypeJSON
		if call.ContentType != "" {
			contentType = call.ContentType
		}
		if chk.ContentType != "" {
			contentType = chk.ContentType
		}

		dataBytes, errDataBytes := ioutil.ReadAll(response.Body)
		if errDataBytes != nil {
			return errors.New("could not read data from response: " + errDataBytes.Error())
		}
		for selector, expect := range chk.Data {
			switch contentType {
			case config.ContentTypeJSON:
				ok, info := check.JSONPath(dataBytes, selector, expect)
				if !ok {
					//log.Println("no match", selector, expect)
					addError(info, ErrorTypeDataMismatch)
				}
			default:
				addError("data contentType: "+contentType+" is not supported (yet?)", ErrorTypeNotImplemented)
			}
		}
	case chk.ContentType != "":
		contentType := response.Header.Get("Content-Type")
		if contentType != chk.ContentType {
			addError("unexpected Content-Type: \""+contentType+"\", expected: \""+chk.ContentType+"\"", ErrorTypeUnexpectedContentType)
		}
	default:
		log.Println("what to check here !?")
	}
	return nil
}
