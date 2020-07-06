package collector

import (
	"github.com/dreadl0ck/petze/watch"
	"testing"
)

func TestCollectorListeners(t *testing.T) {
	var actualResult watch.Result
	c, _ := NewCollector("")
	c.RegisterListener(func(result watch.Result) {
		actualResult = result
	})

	expectedResult := watch.Result{ID: "some-fake-id"}
	c.NotifyListeners(expectedResult)

	if actualResult.ID != expectedResult.ID {
		t.Error("actual result is not equal to the expected result")
	}
}
