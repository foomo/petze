package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	ContentTypeJSON = "application/json"
)

const serverConfigFile = "petze.yml"

type Expect struct {
	Max      *int64
	Min      *int64
	Count    *int64
	Contains string
	Equals   interface{}
}

type Check struct {
	Comment     string
	Data        map[string]Expect
	Goquery     map[string]Expect
	Header      map[string][]string
	Duration    time.Duration
	StatusCode  int64
	ContentType string `yaml:"content-type"`
}

type Call struct {
	URI         string
	URL         string
	Method      string
	Data        interface{}
	ContentType string `yaml:"content-type"`
	Check       []Check
	Header      map[string][]string
}

// Service a service to monitor
type Service struct {
	ID       string
	Endpoint string
	Interval time.Duration
	Session  []Call
}

type Server struct {
	Address       string
	BasicAuthFile string
	TLS *struct {
		Address string
		Cert    string
		Key     string
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
