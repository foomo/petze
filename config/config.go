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
	Max      *int64      `yaml:"max"`
	Min      *int64      `yaml:"min"`
	Count    *int64      `yaml:"count"`
	Contains string      `yaml:"contains"`
	Equals   interface{} `yaml:"equals"`
}

type Check struct {
	Comment     string            `yaml:"comment"`
	JSONPath    map[string]Expect `yaml:"jsonPath"`
	GoQuery     map[string]Expect `yaml:"goQuery"`
	Headers     map[string]string `yaml:"headers"`
	Regex       map[string]Expect `yaml:"regex"`
	Duration    time.Duration     `yaml:"duration"`
	StatusCode  int64             `yaml:"statusCode"`
	ContentType string            `yaml:"contentType"`
	Redirect    string            `yaml:"redirect"`
	MatchReply  string            `yaml:"matchReply"`
}

type Call struct {
	// allow to overwrite scheme
	Scheme      string            `yaml:"scheme"`
	URI         string            `yaml:"uri"`
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`
	Data        interface{}       `yaml:"data"`
	ContentType string            `yaml:"contentType"`
	Check       []Check           `yaml:"check"`
	Headers     map[string]string `yaml:"headers"`
	Comment     string            `yaml:"comment"`
}

// Service is a service to monitor
// The service id is the name of the file including its relative path in the config folder
// e.g:
// google.yml -> service ID: google
// cluster1/service1.yml -> service ID: cluster1/service1
type Service struct {

	// service identifier
	ID       string        `yaml:"id"`
	Endpoint string        `yaml:"endpoint"`
	Interval time.Duration `yaml:"interval"`

	Session    []Call        `yaml:"session"`

	// Notifications
	NotifyIfResolved bool `yaml:"notifyIfResolved"`

	// Generate an error if the TLS certificate will expire in less then
	TLSWarning time.Duration `yaml:"tlsWarning"`
}

// Server models the petze.yml main config file
type Server struct {

	// endpoint to expose metrics
	Address       string `yaml:"address"`

	// auth
	BasicAuthFile string `yaml:"basicAuthFile"`

	// Notifications
	TLS *struct {
		Address string `yaml:"address"`
		Cert    string `yaml:"cert"`
		Key     string `yaml:"key"`
	}

	SMTP *struct {
		User   string   `yaml:"user"`
		Pass   string   `yaml:"pass"`
		Server string   `yaml:"server"`
		Port   int      `yaml:"port"`
		From   string   `yaml:"from"`
		To     []string `yaml:"to"`
	} `yaml:"smtp"`

	Slack string `yaml:"slack"`
	Sms   *SMS   `yaml:"sms"`
}

type SMS struct {
	SendInBlueAPIKey string `yaml:"sendInBlueAPIKey"`

	TwilioSID   string `yaml:"twilioSID"`
	TwilioToken string `yaml:"twilioToken"`

	To   []string `yaml:"to"`
	From string   `yaml:"from"`
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
