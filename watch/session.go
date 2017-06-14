package watch

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"time"

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

		for _, chk := range call.Check {
			ctx := &CheckContext{
				response: response,
				check:    chk,
				call:     call,
				duration: duration,
			}
			r.Errors = checkResponse(ctx)
		}

	}
	return nil
}

type CheckContext struct {
	response *http.Response
	check    config.Check
	call     config.Call
	duration time.Duration
}

func checkResponse(ctx *CheckContext) []Error {
	errs := []Error{}

	dataValidator := &ResponseDataValidator{}
	dataValidator.Validate(ctx)
	switch true {
	case ctx.check.Duration > 0:
		if ctx.duration > ctx.check.Duration {
			errs = append(errs, Error{
				Error: fmt.Sprint("call duration ", ctx.duration, " exceeded ", ctx.check.Duration),
				Type:  ErrorTypeServerTooSlow,
			})
		}
	case ctx.check.Goquery != nil:
		// go query
		doc, errDoc := goquery.NewDocumentFromResponse(ctx.response)
		if errDoc != nil {
			errs = append(errs, Error{
				Error: errDoc.Error(),
				Type:  ErrorTypeGoQuerySyntax,
			})
		} else {
			for selector, expect := range ctx.check.Goquery {
				ok, info := check.Goquery(doc, selector, expect)
				if !ok {
					errs = append(errs, Error{
						Error: info,
						Type:  ErrorTypeGoQueryMismatch,
					})
				}
			}
		}
	case ctx.check.ContentType != "":
		contentType := ctx.response.Header.Get("Content-Type")
		if contentType != ctx.check.ContentType {
			errs = append(errs, Error{
				Error: "unexpected Content-Type: \"" + contentType + "\", expected: \"" + ctx.check.ContentType + "\"",
				Type:  ErrorTypeUnexpectedContentType,
			})
		}
	}
	return errs
}
