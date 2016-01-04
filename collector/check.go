package collector

import (
	"time"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/watch"
)

func serviceIsDown(results []*watch.Result) (down bool, start time.Time) {
	isDown := false
	var downSince time.Time
	for _, result := range results {
		if len(result.Errors) > 0 && isDown == false {
			isDown = true
			downSince = result.Timestamp
		} else if len(result.Errors) == 0 && isDown == true {
			isDown = false
		}
	}
	if isDown {
		return true, downSince
	}

	return false, downSince
}

func runChecks(services map[string]*config.Service, results map[string][]*watch.Result) map[string]map[string]*Alert {
	alerts := make(map[string]map[string]*Alert)
	for serviceID, service := range services {
		serviceResults, ok := results[serviceID]
		if ok {
			serviceAlerts := make(map[string]*Alert)
			// classic is down
			isDown, downStartTime := serviceIsDown(serviceResults)
			if isDown && (int(time.Since(downStartTime).Seconds()) > service.Alert.After) {
				serviceAlerts[AlertTypeDown] = &Alert{
					Type:  AlertTypeDown,
					Start: downStartTime,
				}
			}
			// error rate maybe
			// next type
			if len(serviceAlerts) > 0 {
				alerts[serviceID] = serviceAlerts
			}
		}
	}
	return alerts
}
