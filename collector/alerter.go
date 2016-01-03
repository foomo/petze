package collector

import (
	"log"
	"time"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"
)

// Alert with type and start time
type Alert struct {
	Type  string
	Start time.Time
}

const (
	// AlertTypeDown classic down
	AlertTypeDown = "down"
)

func newAlerter() *alerter {
	return &alerter{
		alerts: make(map[string]map[string]*Alert),
	}
}

// alert people
type alerter struct {
	alerts map[string]map[string]*Alert
}

func (a *alerter) checkServices(services map[string]*config.Service, results map[string][]*watch.Result, people map[string]*config.Person) {
	alerts := runChecks(services, results)
	for serviceID, serviceAlerts := range alerts {
		for alertType, serviceAlertOfAType := range serviceAlerts {
			log.Println("alert for", serviceID, "of type", alertType, ":", serviceAlertOfAType)
		}
	}
}
