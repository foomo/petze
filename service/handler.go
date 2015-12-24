package service

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *server) GETCollectorConfigServices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//fmt.Println(s.collector.UpdatePeople(), s.collector.UpdateServices())
	jsonReply("GETCollectorConfigServices", w)
}
