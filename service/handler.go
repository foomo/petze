package service

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/foomo/petze/watch"
	"github.com/julienschmidt/httprouter"
)

type ServiceStatus struct {
	ID      string         `json:"id"`
	Results []watch.Result `json:"results"`
}

func (s *server) GETServices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonReply("GETCollectorConfigServices", w)
}

func (s *server) GETServicesStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	limitInt := 1000
	limitIntCandidate, err := strconv.Atoi(r.FormValue("limit"))
	if err == nil {
		limitInt = limitIntCandidate
	}
	status := []ServiceStatus{}

	serviceResults := s.collector.GetResults()
	serviceIDs := []string{}
	for serviceID := range serviceResults {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)
	for _, serviceID := range serviceIDs {
		results := serviceResults[serviceID]
		if len(results) > limitInt {
			results = results[len(results)-limitInt:]
		}
		status = append(status, ServiceStatus{
			ID:      serviceID,
			Results: results,
		})
	}
	jsonReply(status, w)
}
