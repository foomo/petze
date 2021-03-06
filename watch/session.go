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

	"io/ioutil"

	"github.com/foomo/petze/config"
)

var userAgent = "Petze Service Monitor"

func SetUserAgentVersion(version string) {
	userAgent += "/"+version
}

func (w *Watcher) runSession(r *Result, client *http.Client) error {

	//log.Println("running session with session length:", len(service.Session))
	//spew.Dump(service)

	endPointURL, errURL := w.service.GetURL()
	if errURL != nil {
		return errors.New("can not run session: " + errURL.Error())
	}
	for indexCall, call := range w.service.Session {

		// copy URL
		callURL := &url.URL{}
		*callURL = *endPointURL

		uriURL, errURIURL := call.GetURL()
		if errURIURL != nil {
			return errURIURL
		}

		callURL.Path = uriURL.Path
		callURL.RawQuery = uriURL.RawQuery

		// overwrite scheme if desired
		if call.Scheme != "" {
			callURL.Scheme = call.Scheme
		}

		call.URL = callURL.String()

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
			//fmt.Println("JSON body", string(dataBytes))
			body = bytes.NewBuffer(dataBytes)
		}

		req, errNewRequest := http.NewRequest(method, callURL.String(), body)
		if errNewRequest != nil {
			return errNewRequest
		}
		start := time.Now()

		// set default user agent first, so it can be overwritten via the custom header fields if desired
		req.Header.Set("User-Agent", userAgent)

		// set the HTTP header fields specified for the call
		for k, v := range call.Headers {
			//fmt.Println("set header", k, v)
			req.Header.Set(k, v)
		}

		// execute the HTTP request
		response, errResponse := client.Do(req)
		if errResponse != nil {
			return errResponse
		}
		defer response.Body.Close()

		// measure time
		duration := time.Since(start)

		// get reader for response body
		responseBodyReader, readerErr := getResponseBodyReader(response)
		if readerErr != nil {
			return readerErr
		}

		// process all checks for the call
		for indexCheck, chk := range call.Check {
			ctx := &CheckContext{
				response:           response,
				responseBodyReader: responseBodyReader,
				check:              chk,
				call:               call,
				duration:           duration,
			}
			for _, newErr := range checkResponse(ctx) {
				newErr.Location = fmt.Sprint("@call[", indexCall, "].check[", indexCheck, "]")
				r.Errors = append(r.Errors, newErr)
			}
			responseBodyReader.Seek(0, io.SeekStart)
		}
	}
	return nil
}

func getResponseBodyReader(response *http.Response) (io.ReadSeeker, error) {
	responseBody, errReadAll := ioutil.ReadAll(response.Body)
	if errReadAll != nil {
		return nil, errors.New("could not read from response" + errReadAll.Error())
	}
	return bytes.NewReader(responseBody), nil
}

type CheckContext struct {
	response           *http.Response
	responseBodyReader io.Reader
	check              config.Check
	call               config.Call
	duration           time.Duration
}

var ContextValidators = []ValidatorFunc{
	ValidateRedirects,
	ValidateHeaders,
	ValidateStatusCode,
	ValidateJsonPath,
	ValidateGoQuery,
	ValidateDuration,
	ValidateContentType,
	ValidateRegex,
	ValidateMatchReply,
}

func checkResponse(ctx *CheckContext) []Error {
	errs := []Error{}

	for _, validator := range ContextValidators {
		errs = append(errs, validator(ctx)...)
	}

	return errs
}
