package watch

import (
	"testing"
	"time"
	"github.com/foomo/petze/config"
	"net/http/httptest"
)

func TestValidateDurationError(t *testing.T) {
	ctx := &CheckContext{
		duration: 200 * time.Millisecond,
		check:    config.Check{Duration: 100 * time.Millisecond },
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
