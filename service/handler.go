package service

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *server) GETServices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//fmt.Println(s.collector.UpdatePeople(), s.collector.UpdateServices())
	jsonReply("GETCollectorConfigServices", w)
}

func (s *server) GETStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonReply(s.collector.GetResults(), w)
}

func (s *server) GETAlerts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jsonReply(s.collector.GetAlerts(), w)
}

func (s *server) POSTUserToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, _ := ioutil.ReadAll(r.Body)
	log.Println("post user token", ps, string(body))
}
