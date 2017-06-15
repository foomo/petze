package collector

import (
	"testing"
	"github.com/foomo/petze/watch"
)

func TestCollectorListeners(t *testing.T) {
	var actualResult watch.Result
	c, _ := NewCollector("")
	c.registerListener(func(result watch.Result) {
		actualResult = result
	})

	expectedResult := watch.Result{ID: "some-fake-id"}
	c.notifyListeners(expectedResult)

	if actualResult.ID != expectedResult.ID {
		t.Error("actual result is not equal to the expected result")
	}
}
