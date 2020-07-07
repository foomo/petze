package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	ContentTypeJSON  = "application/json"
	serverConfigFile = "petze.yml"
)

type Expect struct {
	Max      *int64
	Min      *int64
	Count    *int64
	Contains string
	Equals   interface{}
}

type Check struct {
	Comment     string
	JSONPath    map[string]Expect `yaml:"json-path"`
	Goquery     map[string]Expect
	Headers     map[string]string
	Regex       map[string]Expect
	Duration    time.Duration
	StatusCode  int64  `yaml:"statuscode"` // TODO: unify naming
	ContentType string `yaml:"content-type"`
	Redirect    string
	MatchReply  string `yaml:"match-reply"`
}

type Call struct {
	Scheme      string // allow to overwrite scheme
	URI         string
	URL         string
	Method      string
	Data        interface{}
	ContentType string `yaml:"content-type"`
	Check       []Check
	Headers     map[string]string
	Comment     string
}

// Service a service to monitor
type Service struct {
	ID       string
	Endpoint string
	Interval time.Duration
	// Generate an error if the TLS certificate will expire in less then
	TLSWarning time.Duration `yaml:"tlswarning"`
	Session   []Call
}

type Server struct {
	Address       string
	BasicAuthFile string
	TLS           *struct {
		Address string
		Cert    string
		Key     string
	}
	SMTP *struct {
		User   string
		Pass   string
		Server string
		Port   int
		From   string
		To     string
	}
}

func (s *Service) GetURL() (u *url.URL, e error) {
	return url.Parse(s.Endpoint)
}

func (s *Service) IsValid() (valid bool, err error) {
	valid = false
	_, errURL := s.GetURL()
	if errURL != nil {

		err = errors.New("endpoint is invalid: " + errURL.Error())
		return
	}
	for callIndex, call := range s.Session {
		_, callErr := call.IsValid()
		if callErr != nil {
			err = errors.New(fmt.Sprint("invalid call in session @", callIndex, " : ", callErr))
			return
		}
	}
	valid = true
	return
}
func (c *Call) GetURL() (u *url.URL, e error) {
	return url.Parse(c.URI)
}

func (c *Call) IsValid() (valid bool, err error) {
	valid = true
	_, errURL := c.GetURL()
	if errURL != nil {
		err = errors.New("invalid uri " + c.URI + " : " + errURL.Error())
		return
	}
	return
}
