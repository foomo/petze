package watch

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/dreadl0ck/petze/check"
	"github.com/dreadl0ck/petze/config"
)

type ValidatorFunc func(ctx *CheckContext) (errs []Error)

func ValidateStatusCode(ctx *CheckContext) (errs []Error) {
	// handle status code checks
	if ctx.check.StatusCode != 0 && ctx.response.StatusCode != int(ctx.check.StatusCode) {
		errs = append(errs, Error{
			Error: "unexpected status code: got " + ctx.response.Status + ", expected: " + strconv.FormatInt(ctx.check.StatusCode, 10),
			Type:  ErrorTypeWrongHTTPStatusCode,
		})
	}
	return
}

func ValidateJsonPath(ctx *CheckContext) (errs []Error) {
	if ctx.check.JSONPath != nil {
		contentType := config.ContentTypeJSON
		if ctx.call.ContentType != "" {
			contentType = ctx.call.ContentType
		}
		if ctx.check.ContentType != "" {
			contentType = ctx.check.ContentType
		}

		dataBytes, errDataBytes := ioutil.ReadAll(ctx.responseBodyReader)
		if errDataBytes != nil {
			errs := append(errs, Error{Error: "could not read data from response: " + errDataBytes.Error()})
			return errs
		}

		for selector, expect := range ctx.check.JSONPath {
			switch contentType {
			case config.ContentTypeJSON:
				ok, info := check.JSONPath(dataBytes, selector, expect)
				if !ok {
					errs = append(errs, Error{
						Error: info,
						Type:  ErrorJsonPath,
					})
				}
			default:
				errs = append(errs, Error{
					Error: "data contentType: " + contentType + " is not supported (yet?)",
					Type:  ErrorTypeNotImplemented,
				})
			}
		}
	}
	return
}

func ValidateDuration(ctx *CheckContext) (errs []Error) {
	if ctx.check.Duration > 0 {
		if ctx.duration > ctx.check.Duration {
			errs = append(errs, Error{
				Error: fmt.Sprint("call duration ", ctx.duration, " exceeded ", ctx.check.Duration),
				Type:  ErrorTypeServerTooSlow,
			})
		}
	}
	return
}

func ValidateGoQuery(ctx *CheckContext) (errs []Error) {
	if ctx.check.Goquery != nil {
		// go query

		doc, errDoc := goquery.NewDocumentFromReader(ctx.responseBodyReader)
		if errDoc != nil {
			errs = append(errs, Error{
				Error: errDoc.Error(),
				Type:  ErrorTypeGoQuery,
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
	}
	return
}

func ValidateContentType(ctx *CheckContext) (errs []Error) {
	if ctx.check.ContentType != "" {
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

func ValidateRegex(ctx *CheckContext) (errs []Error) {
	if ctx.check.Regex != nil {
		data, errDataBytes := ioutil.ReadAll(ctx.responseBodyReader)
		if errDataBytes != nil {
			errs = append(errs, Error{Error: "could not read data from response: " + errDataBytes.Error()})
			return
		}
		for regexString, expect := range ctx.check.Regex {
			if ok, info := check.Regex(data, regexString, expect); ok == false {
				errs = append(errs, Error{Error: info, Type: ErrorRegex})
			}
		}
	}
	return
}
