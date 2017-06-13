package watch

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"time"

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

		req, errNewRequest := http.NewRequest(http.MethodGet, callURL.String(), nil)
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
			errCheck := checkResponse(r, response, check, duration)
			if errCheck != nil {
				return errCheck
			}
		}
	}
	return nil
}

func checkResponse(r *Result, response *http.Response, chk config.Check, callDuration time.Duration) error {
	log.Println("checking response", chk)

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
