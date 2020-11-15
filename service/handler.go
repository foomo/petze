package service

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/foomo/petze/watch"
	"github.com/julienschmidt/httprouter"
)

type ServiceStatus struct {
	ID             string                `json:"id"`
	ServiceResults []watch.ServiceResult `json:"results"`
}

type HostStatus struct {
	ID          string             `json:"id"`
	HostResults []watch.HostResult `json:"results"`
}

func (s *server) GETServices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonReply("GETCollectorConfigServices", w)
}

func (s *server) GETHosts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonReply("GETCollectorConfigHosts", w)
}

func (s *server) GETServicesStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	limitInt := 1000
	limitIntCandidate, err := strconv.Atoi(r.FormValue("limit"))
	if err == nil {
		limitInt = limitIntCandidate
	}
	status := []ServiceStatus{}

	serviceResults := s.collector.GetServiceResults()
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
			ID:             serviceID,
			ServiceResults: results,
		})
	}
	jsonReply(status, w)
}

func (s *server) GETHostsStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	limitInt := 1000
	limitIntCandidate, err := strconv.Atoi(r.FormValue("limit"))
	if err == nil {
		limitInt = limitIntCandidate
	}
	status := []HostStatus{}

	serviceResults := s.collector.GetHostResults()
	hostIDs := []string{}
	for hostID := range serviceResults {
		hostIDs = append(hostIDs, hostID)
	}
	sort.Strings(hostIDs)
	for _, hostID := range hostIDs {
		results := serviceResults[hostID]
		if len(results) > limitInt {
			results = results[len(results)-limitInt:]
		}
		status = append(status, HostStatus{
			ID:          hostID,
			HostResults: results,
		})
	}
	jsonReply(status, w)
}
