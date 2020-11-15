package collector

import (
	"encoding/json"
	"time"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"

	log "github.com/sirupsen/logrus"
)

type ServiceResultListener func(watch.ServiceResult)
type HostResultListener func(watch.HostResult)

// Collector collects stats on hosts & services
type Collector struct {
	servicesConfigDir      string
	hostsConfigDir         string
	chanServices           chan map[string]*config.Service
	chanHosts              chan map[string]*config.Host
	chanGetServiceResults  chan map[string][]watch.ServiceResult
	chanGetHostResults     chan map[string][]watch.HostResult
	serviceWatchers        map[string]*watch.ServiceWatcher
	hostWatchers           map[string]*watch.HostWatcher
	serviceResultListeners []ServiceResultListener
	hostResultListeners    []HostResultListener
	services               map[string]*config.Service
	hosts                  map[string]*config.Host
}

// NewCollector construct a collector - it will watch its config files for changes
func NewCollector(servicesConfigDir string, hostsConfigDir string) (c *Collector, err error) {
	c = &Collector{
		servicesConfigDir:      servicesConfigDir,
		hostsConfigDir:         hostsConfigDir,
		services:               make(map[string]*config.Service),
		hosts:                  make(map[string]*config.Host),
		chanServices:           make(chan map[string]*config.Service),
		chanHosts:              make(chan map[string]*config.Host),
		chanGetServiceResults:  make(chan map[string][]watch.ServiceResult),
		chanGetHostResults:     make(chan map[string][]watch.HostResult),
		serviceWatchers:        make(map[string]*watch.ServiceWatcher),
		hostWatchers:           make(map[string]*watch.HostWatcher),
		serviceResultListeners: make([]ServiceResultListener, 0),
		hostResultListeners:    make([]HostResultListener, 0),
	}

	return c, nil
}

// Starts collection of results and configuration watch
func (c *Collector) Start() {
	go c.collect()
	go c.configWatch()
}

const maxServiceResults = 1000
const maxHostResults = 1000

func (c *Collector) RegisterServiceListener(listener ServiceResultListener) {
	c.serviceResultListeners = append(c.serviceResultListeners, listener)
}

func (c *Collector) RegisterHostListener(listener HostResultListener) {
	c.hostResultListeners = append(c.hostResultListeners, listener)
}

func (c *Collector) NotifyServiceListeners(result watch.ServiceResult) {
	for _, listener := range c.serviceResultListeners {
		listener(result)
	}
}

func (c *Collector) NotifyHostListeners(result watch.HostResult) {
	for _, listener := range c.hostResultListeners {
		listener(result)
	}
}

func (c *Collector) collect() {

	chanServiceResult := make(chan watch.ServiceResult)
	chanHostResult := make(chan watch.HostResult)
	serviceResults := map[string][]watch.ServiceResult{}
	hostResults := map[string][]watch.HostResult{}

	for {
		select {
		case <-c.chanGetServiceResults:
			serviceResultsCopy := map[string][]watch.ServiceResult{}
			for name, serviceResults := range serviceResults {
				serviceResultsCopy[name] = serviceResults
			}
			c.chanGetServiceResults <- serviceResultsCopy

		case <-c.chanGetHostResults:
			hostResultsCopy := map[string][]watch.HostResult{}
			for name, hostResults := range hostResults {
				hostResultsCopy[name] = hostResults
			}
			c.chanGetHostResults <- hostResultsCopy
		case newServices := <-c.chanServices:
			c.services = newServices

			var lastErrors = make(map[string][]watch.Error)

			// stop old watchers
			for oldWatcherID, oldWatcher := range c.serviceWatchers {
				oldWatcher.Watcher.Stop()

				// if the service had errors before updating the config
				// store them in a map so we can transfer them to the updated watchers
				if len(oldWatcher.Watcher.LastErrors()) > 0 {
					lastErrors[oldWatcherID] = oldWatcher.Watcher.LastErrors()
				}

				delete(c.serviceWatchers, oldWatcherID)
			}
			// setup new watchers
			for serviceID, service := range c.services {
				// check if the service had errors before being updated
				lastErrs, ok := lastErrors[serviceID]
				if ok {
					// transfer errors to new watcher
					newWatcher := watch.WatchService(service, chanServiceResult, chanHostResult, c.hosts)
					newWatcher.SetLastErrors(lastErrs)
					c.serviceWatchers[serviceID] = newWatcher
				} else {
					// no errors - init a new watcher
					// TODO: when starting up c.hosts is empty...
					c.serviceWatchers[serviceID] = watch.WatchService(service, chanServiceResult, chanHostResult, c.hosts)
				}
				// reset stored results
				_, ok = serviceResults[serviceID]
				if !ok {
					serviceResults[serviceID] = []watch.ServiceResult{}
				}
			}

		case newHosts := <-c.chanHosts:
			c.hosts = newHosts

			var lastErrors = make(map[string][]watch.Error)

			// stop old watchers
			for oldWatcherID, oldWatcher := range c.hostWatchers {
				oldWatcher.Watcher.Stop()

				// if the host had errors before updating the config
				// store them in a map so we can transfer them to the updated watchers
				if len(oldWatcher.Watcher.LastErrors()) > 0 {
					lastErrors[oldWatcherID] = oldWatcher.Watcher.LastErrors()
				}

				delete(c.hostWatchers, oldWatcherID)
			}

			// setup new watchers
			for hostID, host := range c.hosts {
				// check if the host had errors before being updated
				lastErrs, ok := lastErrors[hostID]
				if ok {
					// transfer errors to new watcher
					newWatcher := watch.WatchHost(host, chanHostResult)
					newWatcher.SetLastErrors(lastErrs)
					c.hostWatchers[hostID] = newWatcher
				} else {
					// no errors - init a new watcher
					c.hostWatchers[hostID] = watch.WatchHost(host, chanHostResult)
				}
				// reset stored results
				_, ok = hostResults[hostID]
				if !ok {
					hostResults[hostID] = []watch.HostResult{}
				}
			}
			// clean up results
			for possiblyUnknownHostID := range hostResults {
				_, ok := c.hostWatchers[possiblyUnknownHostID]
				if !ok {
					// clean up results
					delete(hostResults, possiblyUnknownHostID)
				}
			}
		case serviceResult := <-chanServiceResult:
			serviceResultsFromChan, ok := serviceResults[serviceResult.Result.ID]
			if ok {
				serviceResultsFromChan = append(serviceResultsFromChan, serviceResult)
				if len(serviceResultsFromChan) > maxServiceResults {
					serviceResultsFromChan = serviceResultsFromChan[len(serviceResultsFromChan)-maxServiceResults:]
				}
				serviceResults[serviceResult.Result.ID] = serviceResultsFromChan

				c.NotifyServiceListeners(serviceResult)
			}
		case hostResult := <-chanHostResult:
			hostResultsFromChan, ok := hostResults[hostResult.Result.ID]
			if ok {
				hostResultsFromChan = append(hostResultsFromChan, hostResult)
				if len(hostResultsFromChan) > maxHostResults {
					hostResultsFromChan = hostResultsFromChan[len(hostResultsFromChan)-maxHostResults:]
				}
				hostResults[hostResult.Result.ID] = hostResultsFromChan

				c.NotifyHostListeners(hostResult)
			}
		}
	}
}

func (c *Collector) GetServiceResults() map[string][]watch.ServiceResult {
	c.chanGetServiceResults <- nil
	return <-c.chanGetServiceResults
}

func (c *Collector) GetHostResults() map[string][]watch.HostResult {
	c.chanGetHostResults <- nil
	return <-c.chanGetHostResults
}

func hashServiceConfig(config map[string]*config.Service) (hash string) {
	hash = "invalid config"
	jsonBytes, errJSON := json.Marshal(config)
	if errJSON == nil {
		hash = string(jsonBytes)
	}
	return hash
}

func hashHostConfig(config map[string]*config.Host) (hash string) {
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
			log.Error("could not read configuration: ", errServices)
		}
		if errServices == nil {
			newHash := hashServiceConfig(services)
			oldHash := hashServiceConfig(c.services)
			if newHash != oldHash {
				log.Info("service configuration update successful")
				c.updateServices()
			}
		}

		hosts, errHosts := config.LoadHosts(c.hostsConfigDir)

		if errHosts != nil {
			log.Error("could not read configuration: ", errHosts)
		}
		if errHosts == nil {
			newHash := hashHostConfig(hosts)
			oldHash := hashHostConfig(c.hosts)
			if newHash != oldHash {
				log.Info("host configuration update successful")
				c.updateHosts()
			}
		}

		// check if service defined in host config exists
		for _, hostConfig := range hosts {
			// validate services
			for _, service := range hostConfig.Services {
				if _, ok := services[service]; !ok {
					log.Fatal("host " + hostConfig.Hostname + " has service " + service + " in its list of services but the service doesn't exist in the service config dir")
				}
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

func (c *Collector) updateHosts() error {
	hosts, err := config.LoadHosts(c.hostsConfigDir)
	if err == nil {
		c.chanHosts <- hosts
	} else {
		log.Warn("could not update hosts:", err)
	}
	return err
}
