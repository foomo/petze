package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"path/filepath"

	"strings"

	yaml "gopkg.in/yaml.v1"
)

const serverConfigFile = "petze.yml"

type Check struct {
	OK bool
}

// Service a service to monitor
type Service struct {
	Endpoint   string
	ID         string
	Interval   int
	MaxRuntime int
	Checks     []Check
}

type Server struct {
	Address       string
	BasicAuthFile string
	TLS           *struct {
		Address string
		Cert    string
		Key     string
	}
}

func LoadServices(configDir string) (services map[string]*Service, err error) {
	services = make(map[string]*Service)
	errLoadServices := loadServicesFromDir(configDir, services)
	if errLoadServices != nil {
		err = errors.New("could not load service configurations from config dir : " + configDir + ",  : " + errLoadServices.Error())
		return
	}
	for id, service := range services {
		service.ID = id
		if service.MaxRuntime == 0 {
			service.MaxRuntime = 1000
		}
		if service.Interval == 0 {
			service.Interval = 60
		}
	}
	return services, nil
}

func LoadServer(configDir string) (server *Server, err error) {
	server = &Server{}
	return server, load(path.Join(configDir, serverConfigFile), &server)
}

func loadServicesFromDir(configDir string, targets map[string]*Service) error {
	absoluteConfigDir, errAbsoluteConfigDir := filepath.Abs(configDir)
	if errAbsoluteConfigDir != nil {
		return errAbsoluteConfigDir
	}
	return filepath.Walk(absoluteConfigDir, func(fp string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(fp, ".yml") {
			p := strings.TrimSuffix(strings.TrimPrefix(fp, absoluteConfigDir+string(os.PathSeparator)), ".yml")
			// fmt.Println(fp, info.Name(), p)
			serviceConfig := &Service{}
			targets[p] = serviceConfig
			return load(fp, &serviceConfig)
		}
		return nil
	})
}

// Load load config from a file
func load(configFile string, target interface{}) error {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	yamlErr := yaml.Unmarshal(configBytes, target)
	if yamlErr != nil {
		return errors.New("could not unmarshal yaml file " + configFile + " : " + yamlErr.Error())
	}
	return nil
}
