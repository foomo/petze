package collector

import (
	"testing"

	"github.com/foomo/petze/watch"
)

func TestCollectorListeners(t *testing.T) {
	var actualResult watch.ServiceResult
	c, _ := NewCollector("", "")
	c.RegisterServiceListener(func(result watch.ServiceResult) {
		actualResult = result
	})

	expectedResult := watch.ServiceResult{
		Result: watch.Result{
			ID: "some-fake-id",
		},
	}
	c.NotifyServiceListeners(expectedResult)

	if actualResult.ID != expectedResult.ID {
		t.Error("actual result is not equal to the expected result")
	}
}
