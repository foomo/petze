package watch

import (
	"fmt"
	"net/http"
	"time"

	"github.com/foomo/petze/config"
)

type Result struct {
	ID         string
	Error      error
	DNS        bool
	Timestamp  time.Time
	RunTime    time.Duration
	StatusCode int
}

type Watcher struct {
	active  bool
	service *config.Service
}

// Watch create a watcher and start watching
func Watch(service *config.Service, chanResult chan *Result) *Watcher {
	w := &Watcher{
		active:  true,
		service: service,
	}
	go w.watchLoop(chanResult)
	return w
}

// Stop watching - beware this is async
func (w *Watcher) Stop() {
	w.active = false
}

func (w *Watcher) watchLoop(chanResult chan *Result) {
	for w.active {
		r := watch(w.service)
		if w.active {
			chanResult <- r
			time.Sleep(time.Second * time.Duration(w.service.Interval))
		}
	}
}

// actual watch
func watch(service *config.Service) (r *Result) {
	r = &Result{
		ID:        service.ID,
		Timestamp: time.Now(),
	}
	response, err := http.DefaultClient.Get(service.Endpoint)
	r.RunTime = time.Since(r.Timestamp)
	if err != nil {
		r.Error = err
		return
	}
	r.StatusCode = response.StatusCode
	if response.StatusCode != http.StatusOK {
		r.Error = fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	return
}
