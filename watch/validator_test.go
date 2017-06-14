package watch

import (
	"github.com/foomo/petze/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type validationCheck struct {
	errorType ErrorType
	length    int
	message   string
}

func TestValidateDurationError(t *testing.T) {
	ctx := &CheckContext{
		duration: 200 * time.Millisecond,
		check:    config.Check{Duration: 100 * time.Millisecond},
	}

	errs := ValidateDuration(ctx)
	if len(errs) != 1 || errs[0].Type != ErrorTypeServerTooSlow {
		t.Fail()
	}
}

func TestValidateContentTypeError(t *testing.T) {
	resp := httptest.NewRecorder()
	resp.Header().Set("Content-Type", "application/xml")
	ctx := &CheckContext{
		response: resp.Result(),
		check:    config.Check{ContentType: "application/json"},
	}

	errs := ValidateContentType(ctx)
	if len(errs) != 1 || errs[0].Type != ErrorTypeUnexpectedContentType {
		t.Fail()
	}
}

func TestValidateGoQueryBadResponseBody(t *testing.T) {
	resp := httptest.NewRecorder()

	ctx := &CheckContext{
		response: resp.Result(),
		check: config.Check{
			Goquery: map[string]config.Expect{},
		},
	}

	errs := ValidateGoQuery(ctx)
	if len(errs) != 1 || errs[0].Type != ErrorTypeGoQuery {
		t.Fail()
	}
}





var validateJsonPathTests = []struct {
	in  *CheckContext
	out validationCheck
}{
	{&CheckContext{
		response: createResponse("{\"hello\":\"world\"}", "application/json"),
		check: config.Check{Data: map[string]config.Expect{"$.hello": {Equals: "world"}}},
	}, validationCheck{"", 0, "failed valid jquery path"}},
}

func createResponse(data, contentType string) *http.Response {
	resp := httptest.NewRecorder()
	resp.HeaderMap.Set("Content-Type", contentType)
	resp.Body.Write([]byte(data))
	return resp.Result()
}

func TestValidateJsonPath(t *testing.T) {
	for _, test := range validateJsonPathTests {
		errs := ValidateJsonPath(test.in)
		if len(errs) != test.out.length || errs[0].Type != test.out.errorType {
			t.Error(test.out.message)
		}
	}
}
