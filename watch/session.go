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
			ctx := &ResponseContext{
				result:   r,
				response: response,
				check:    check,
				call:     call,
				duration: duration,
			}
			errCheck := checkResponse(ctx)
			if errCheck != nil {
				return errCheck
			}
		}

	}
	return nil
}

type ResponseContext struct {
	result   *Result
	response *http.Response
	check    config.Check
	call     config.Call
	duration time.Duration
}

func checkResponse(ctx *ResponseContext) error {
	addError := func(err string, t ErrorType) {
		ctx.result.addError(errors.New(err), t, ctx.check.Comment)
	}
	switch true {
	case ctx.check.Duration > 0:
		if ctx.duration > ctx.check.Duration {
			addError(fmt.Sprint("call duration ", ctx.duration, " exceeded ", ctx.check.Duration), ErrorTypeServerTooSlow)
		}
	case ctx.check.Goquery != nil:
		// go query
		doc, errDoc := goquery.NewDocumentFromResponse(ctx.response)
		if errDoc != nil {
			return errDoc
		}
		for selector, expect := range ctx.check.Goquery {
			ok, info := check.Goquery(doc, selector, expect)
			if !ok {
				//log.Println("no match", selector, expect)
				addError(info, ErrorTypeGoQueryMismatch)
			}
		}
	case ctx.check.Data != nil:
		contentType := config.ContentTypeJSON
		if ctx.call.ContentType != "" {
			contentType = ctx.call.ContentType
		}
		if ctx.check.ContentType != "" {
			contentType = ctx.check.ContentType
		}

		dataBytes, errDataBytes := ioutil.ReadAll(ctx.response.Body)
		if errDataBytes != nil {
			return errors.New("could not read data from response: " + errDataBytes.Error())
		}
		for selector, expect := range ctx.check.Data {
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
	case ctx.check.ContentType != "":
		contentType := ctx.response.Header.Get("Content-Type")
		if contentType != ctx.check.ContentType {
			addError("unexpected Content-Type: \""+contentType+"\", expected: \""+ctx.check.ContentType+"\"", ErrorTypeUnexpectedContentType)
		}
	default:
		log.Println("what to check here !?")
	}
	return nil
}
