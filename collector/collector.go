package collector

import (
	"encoding/json"
	"log"
	"time"

	"fmt"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"
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
			// stop old watchers
			for oldWatcherID, oldWatcher := range c.watchers {
				oldWatcher.Stop()
				delete(c.watchers, oldWatcherID)
			}
			// setup new watches
			for serviceID, service := range c.services {
				c.watchers[serviceID] = watch.Watch(service, chanResult)
				_, ok := results[serviceID]
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
			log.Println("could not read configuration:", errServices)
		}
		if errServices == nil {
			newHash := hashServiceConfig(services)
			oldHash := hashServiceConfig(c.services)
			if newHash != oldHash {
				fmt.Println("there was a successful configuration update")
				fmt.Println(oldHash)
				fmt.Println(newHash)
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
		log.Println("could not update services:", err)
	}
	return err
}
