package collector

import (
	"log"
	"os"
	"time"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"
)

// Collector collects stats on services
type Collector struct {
	servicesConfigfile string
	peopleConfigfile   string
	chanPeople         chan map[string]*config.Person
	chanServices       chan map[string]*config.Service
	watchers           map[string]*watch.Watcher
	results            map[string][]*watch.Result
	services           map[string]*config.Service
	alerter            *alerter
}

// NewCollector construct a collector - it will watch its config files for changes
func NewCollector(servicesConfigfile string, peopleConfigfile string) (c *Collector, err error) {
	c = &Collector{
		servicesConfigfile: servicesConfigfile,
		services:           make(map[string]*config.Service),
		peopleConfigfile:   peopleConfigfile,
		chanPeople:         make(chan map[string]*config.Person),
		chanServices:       make(chan map[string]*config.Service),
		watchers:           make(map[string]*watch.Watcher),
		results:            make(map[string][]*watch.Result),
	}
	c.alerter = newAlerter()
	go c.collect()
	go c.configWatch()
	return c, nil
}

const maxResults = 1000

func (c *Collector) collect() {

	people := make(map[string]*config.Person)
	//services := make(map[string]*config.Service)
	chanResult := make(chan *watch.Result)

	for {
		select {
		case newPeople := <-c.chanPeople:
			people = newPeople
			//log.Println("updated people", people)
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
				_, ok := c.results[serviceID]
				if !ok {
					c.results[serviceID] = []*watch.Result{}
				}
			}
			// clean up results
			for possiblyUnknownServiceID := range c.results {
				_, ok := c.watchers[possiblyUnknownServiceID]
				if !ok {
					// clean up results
					delete(c.results, possiblyUnknownServiceID)
				}
			}
		case result := <-chanResult:
			results, ok := c.results[result.ID]
			if ok {
				results = append(results, result)
				if len(results) > maxResults {
					results = results[len(results)-maxResults:]
				}
				c.results[result.ID] = results
			}
			c.alerter.checkServices(c.services, c.results, people)
		case <-time.After(time.Second * 10):
			c.alerter.checkServices(c.services, c.results, people)
		}
	}
}

// GetResults get current results
func (c *Collector) GetResults() map[string][]*watch.Result {
	return c.results
}

// GetAlerts get current alerts
func (c *Collector) GetAlerts() map[string]map[string]*Alert {
	return runChecks(c.services, c.results)
}

func getLastChange(filename string) int64 {
	c := int64(0)
	info, err := os.Stat(filename)
	if err == nil {
		c = info.ModTime().UnixNano()
	}
	return c
}

func (c *Collector) configWatch() {
	serviceLastChange := int64(0)
	peopleLastChange := int64(0)
	for {
		newServiceLastChange := getLastChange(c.servicesConfigfile)
		newPeopleLastChange := getLastChange(c.peopleConfigfile)
		if newPeopleLastChange > peopleLastChange {
			c.updatePeople()
			peopleLastChange = newPeopleLastChange
		}
		if newServiceLastChange > serviceLastChange {
			c.updateServices()
			serviceLastChange = newServiceLastChange
		}
		time.Sleep(time.Second)
	}
}

func (c *Collector) updatePeople() error {
	people, err := config.LoadPeople(c.peopleConfigfile)
	if err == nil {
		c.chanPeople <- people
	} else {
		log.Println("could not update people", err)
	}
	return err
}

func (c *Collector) updateServices() error {
	services, err := config.LoadServices(c.servicesConfigfile)
	if err == nil {
		c.chanServices <- services
	} else {
		log.Println("could not update services", err)
	}
	return err
}
