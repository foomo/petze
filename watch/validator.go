package watch

import (
	"io/ioutil"
	"github.com/foomo/petze/check"
)

import (
	"github.com/foomo/petze/config"
)

type Validator interface {
	Validate(ctx *CheckContext) []Error
}

type ResponseDataValidator struct {
}

func (ResponseDataValidator) Validate(ctx *CheckContext) []Error {
	errs := []Error{}
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
			errs := append(errs, Error{Error: "could not read data from response: " + errDataBytes.Error() })
			return errs
		}

		for selector, expect := range ctx.check.Data {
			switch contentType {
			case config.ContentTypeJSON:
				ok, _ := check.JSONPath(dataBytes, selector, expect)
				if !ok {
					errs = append(errs, Error{
						Error: "could not read data from response: " + errDataBytes.Error(),
						Type:  ErrorTypeDataMismatch,
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

	return errs
}
