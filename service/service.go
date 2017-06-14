package service

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"

	auth "github.com/abbot/go-http-auth"
	"github.com/foomo/petze/collector"
	"github.com/foomo/petze/config"
	"github.com/julienschmidt/httprouter"
)

func jsonReply(data interface{}, w http.ResponseWriter) error {
	jsonBytes, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
	return nil
}

type server struct {
	router    *httprouter.Router
	collector *collector.Collector
}

func newServer(servicesConfigfile string) (s *server, err error) {
	coll, err := collector.NewCollector(servicesConfigfile)
	s = &server{
		router:    httprouter.New(),
		collector: coll,
	}
	s.router.GET("/services", s.GETServices)
	s.router.GET("/status", s.GETStatus)
	return s, nil
}

type basicAuthHandler struct {
	server        *server
	authenticator *auth.BasicAuth
}

func newBasicAuthHandler(server *server, htpasswordFile string) (ba *basicAuthHandler) {
	var authenticator *auth.BasicAuth

	if htpasswordFile != "" {
		secretProvider := auth.HtpasswdFileProvider(htpasswordFile)
		authenticator = auth.NewBasicAuthenticator("petze", secretProvider)
	}

	return &basicAuthHandler{
		server:        server,
		authenticator: authenticator,
	}
}

func (ba *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if ba.authenticator != nil {
		user := ba.authenticator.CheckAuth(r)
		if len(user) == 0 {
			ba.authenticator.RequireAuth(w, r)
			return
		}
	}

	ba.server.router.ServeHTTP(w, r)
}

func getTLSConfig() *tls.Config {
	c := &tls.Config{}
	c.MinVersion = tls.VersionTLS12
	c.PreferServerCipherSuites = true
	c.CipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	}
	c.CurvePreferences = []tls.CurveID{
		tls.CurveP256,
		tls.CurveP384,
		tls.CurveP521,
	}
	return c
}

// Run as a server
func Run(c *config.Server, servicesConfigfile string) error {
	s, err := newServer(servicesConfigfile)
	if err != nil {
		return err
	}
	log.Println("starting petze server on: ", c.Address)

	if c.BasicAuthFile != "" {
		log.Println("\t using basic auth from: ", c.BasicAuthFile)
	}

	ba := newBasicAuthHandler(s, c.BasicAuthFile)

	errorChan := make(chan (error))
	if len(c.Address) > 0 {
		go func() {
			errorChan <- http.ListenAndServe(c.Address, ba)
		}()
	}

	if c.TLS != nil {
		go func() {
			log.Println("tls is configured: ", c.TLS)
			tlsServer := &http.Server{
				Addr:      c.TLS.Address,
				Handler:   ba,
				TLSConfig: getTLSConfig(),
			}
			errorChan <- tlsServer.ListenAndServeTLS(c.TLS.Cert, c.TLS.Key)
		}()
	}
	return <-errorChan
}
