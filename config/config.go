package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Alert struct {
	After int
}

type Contact struct {
	Email string
	Phone string
}

type Person struct {
	Name    string
	Contact Contact
}

// Service a service to monitor
type Service struct {
	Endpoint string
	ID       string
	Interval int
	Alert    *Alert
}

type Server struct {
	APN struct {
		Gateway string
		Pemfile string
	}
	Address       string
	BasicAuthFile string
	TLS           *struct {
		Address string
		Cert    string
		Key     string
	}
}

func LoadPeople(configFile string) (people map[string]*Person, err error) {
	people = make(map[string]*Person)
	return people, load(configFile, &people)
}

func LoadServices(configFile string) (services map[string]*Service, err error) {
	services = make(map[string]*Service)
	err = load(configFile, &services)
	if err != nil {
		return
	}
	for id, service := range services {
		service.ID = id
		if service.Interval == 0 {
			service.Interval = 60
		}
		if service.Alert == nil {
			service.Alert = &Alert{
				After: 300,
			}
		}
	}
	return services, nil
}

func LoadServer(configFile string) (server *Server, err error) {
	server = &Server{}
	return server, load(configFile, &server)
}

// Load load config from a file
func load(configFile string, target interface{}) error {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(configBytes, target)
}
