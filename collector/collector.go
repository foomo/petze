package collector

import (
	"encoding/json"
	"time"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"

	log "github.com/sirupsen/logrus"
)

type ResultListener func(watch.Result)

// Collector collects stats on services
type Collector struct {
	servicesConfigDir string
	chanServices      chan map[string]*config.Service
	chanGetResults    chan map[string][]watch.Result
	watchers          map[string]*watch.Watcher
	resultListeners   []ResultListener
	services          map[string]*config.Service
}

// NewCollector construct a collector - it will watch its config files for changes
func NewCollector(servicesConfigDir string) (c *Collector, err error) {
	c = &Collector{
		servicesConfigDir: servicesConfigDir,
		services:          make(map[string]*config.Service),
		chanServices:      make(chan map[string]*config.Service),
		chanGetResults:    make(chan map[string][]watch.Result),
		watchers:          make(map[string]*watch.Watcher),
		resultListeners:   make([]ResultListener, 0),
	}

	return c, nil
}

// Starts collection of results and configuration watch
func (c *Collector) Start() {
	go c.collect()
	go c.configWatch()
}

const maxResults = 1000

func (c *Collector) RegisterListener(listener ResultListener) {
	c.resultListeners = append(c.resultListeners, listener)
}

func (c *Collector) NotifyListeners(result watch.Result) {
	for _, listener := range c.resultListeners {
		listener(result)
	}
}

func (c *Collector) collect() {

	chanResult := make(chan watch.Result)
	results := map[string][]watch.Result{}

	for {
		select {
		case <-c.chanGetResults:
			resultsCopy := map[string][]watch.Result{}
			for name, results := range results {
				resultsCopy[name] = results
			}
			c.chanGetResults <- resultsCopy
		case newServices := <-c.chanServices:
			c.services = newServices

			var lastErrors = make(map[string][]watch.Error)

			// stop old watchers
			for oldWatcherID, oldWatcher := range c.watchers {
				oldWatcher.Stop()

				// if the service had errors before updating the config
				// store them in a map so we can transfer them to the updated watchers
				if len(oldWatcher.LastErrors()) > 0 {
					lastErrors[oldWatcherID] = oldWatcher.LastErrors()
				}

				delete(c.watchers, oldWatcherID)
			}

			// setup new watchers
			for serviceID, service := range c.services {
				// check if the service had errors before being updated
				lastErrs, ok := lastErrors[serviceID]
				if ok {
					// transfer errors to new watcher
					newWatcher := watch.Watch(service, chanResult)
					newWatcher.SetLastErrors(lastErrs)
					c.watchers[serviceID] = newWatcher
				} else {
					// no errors - init a new watcher
					c.watchers[serviceID] = watch.Watch(service, chanResult)
				}
				// reset stored results
				_, ok = results[serviceID]
				if !ok {
					results[serviceID] = []watch.Result{}
				}
			}
			// clean up results
			for possiblyUnknownServiceID := range results {
				_, ok := c.watchers[possiblyUnknownServiceID]
				if !ok {
					// clean up results
					delete(results, possiblyUnknownServiceID)
				}
			}
		case result := <-chanResult:
			serviceResults, ok := results[result.ID]
			if ok {
				serviceResults = append(serviceResults, result)
				if len(serviceResults) > maxResults {
					serviceResults = serviceResults[len(serviceResults)-maxResults:]
				}
				results[result.ID] = serviceResults

				c.NotifyListeners(result)
			}
		}
	}
}

// GetResults get current results
func (c *Collector) GetResults() map[string][]watch.Result {
	c.chanGetResults <- nil
	return <-c.chanGetResults
}

func hashServiceConfig(config map[string]*config.Service) (hash string) {
	hash = "invalid config"
	jsonBytes, errJSON := json.Marshal(config)
	if errJSON == nil {
		hash = string(jsonBytes)
	}
	return hash
}

func (c *Collector) configWatch() {
	for {
		services, errServices := config.LoadServices(c.servicesConfigDir)
		if errServices != nil {
			log.Error("could not read configuration:", errServices)
		}
		if errServices == nil {
			newHash := hashServiceConfig(services)
			oldHash := hashServiceConfig(c.services)
			if newHash != oldHash {
				log.Info("configuration update successful")
				c.updateServices()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (c *Collector) updateServices() error {
	services, err := config.LoadServices(c.servicesConfigDir)
	if err == nil {
		c.chanServices <- services
	} else {
		log.Warn("could not update services:", err)
	}
	return err
}
