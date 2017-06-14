package watch

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/foomo/petze/check"
	"github.com/foomo/petze/config"
	"io/ioutil"
)

type ValidatorFunc func(ctx *CheckContext) (errs []Error)

func ValidateJsonPath(ctx *CheckContext) (errs []Error) {
	if ctx.check.Data != nil {
		contentType := config.ContentTypeJSON
		if ctx.call.ContentType != "" {
			contentType = ctx.call.ContentType
		}
		if ctx.check.ContentType != "" {
			contentType = ctx.check.ContentType
		}

		dataBytes, errDataBytes := ioutil.ReadAll(ctx.response.Body)
		if errDataBytes != nil {
			errs := append(errs, Error{Error: "could not read data from response: " + errDataBytes.Error()})
			return errs
		}

		for selector, expect := range ctx.check.Data {
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
		doc, errDoc := goquery.NewDocumentFromResponse(ctx.response)
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
