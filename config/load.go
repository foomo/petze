package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// warn one week before the cert will expire by default
const defaultTLSExpiryWarning = 7 * 24 * time.Hour

func LoadServices(configDir string) (services map[string]*Service, err error) {
	services = make(map[string]*Service)
	errLoadServices := loadServicesFromDir(configDir, services)
	if errLoadServices != nil {
		err = errors.New("could not load service configurations from config dir : " + configDir + ",  : " + errLoadServices.Error())
		return
	}
	for id, service := range services {
		service.ID = id
		if service.Interval == 0 {
			service.Interval = 60
		}
	}
	return services, nil
}

func LoadHosts(configDir string, services map[string]*Service) (hosts map[string]*Host, err error) {
	hosts = make(map[string]*Host)
	errLoadServices := loadHostsFromDir(configDir, hosts, services)
	if errLoadServices != nil {
		err = errors.New("could not load host configurations from config dir : " + configDir + ",  : " + errLoadServices.Error())
		return
	}
	for id, host := range hosts {
		host.ID = id
		if host.Interval == 0 {
			host.Interval = 60
		}
	}
	return hosts, nil
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
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(fp, ".yml") && info.Name() != "petze.yml" {
			p := strings.TrimSuffix(strings.TrimPrefix(fp, absoluteConfigDir+string(os.PathSeparator)), ".yml")
			serviceConfig := &Service{}
			targets[p] = serviceConfig
			loadErr := load(fp, &serviceConfig)
			if loadErr != nil {
				return loadErr
			}
			for i, call := range serviceConfig.Session {
				if call.Data != nil {
					serviceConfig.Session[i].Data = fixYamlMapsForJSON(call.Data, 0)
				}
			}
			if serviceConfig.TLSWarning == 0 {
				serviceConfig.TLSWarning = defaultTLSExpiryWarning
			}
			return nil
		}
		return nil
	})
}

func loadHostsFromDir(configDir string, targets map[string]*Host, services map[string]*Service) error {
	absoluteConfigDir, errAbsoluteConfigDir := filepath.Abs(configDir)
	if errAbsoluteConfigDir != nil {
		return errAbsoluteConfigDir
	}
	return filepath.Walk(absoluteConfigDir, func(fp string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") && strings.HasSuffix(fp, ".yml") && info.Name() != "petze.yml" {
			p := strings.TrimSuffix(strings.TrimPrefix(fp, absoluteConfigDir+string(os.PathSeparator)), ".yml")
			hostConfig := &Host{}
			targets[p] = hostConfig
			loadErr := load(fp, &hostConfig)
			if loadErr != nil {
				return loadErr
			}

			// check if services exist
			for _, hostService := range hostConfig.Services {
				// validate services
				_, hostServiceHasConfig := services[hostService]
				if !hostServiceHasConfig {
					return errors.New("host " + hostConfig.DomainName + " has service " + hostService + " in its list of services but the service doesn't exist in the service config dir")
				}
			}
			return nil
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
	yamlErr := yaml.UnmarshalStrict(configBytes, target)
	if yamlErr != nil {
		return errors.New("could not unmarshal yaml file " + configFile + " : " + yamlErr.Error())
	}
	return nil
}

func fixYamlMapsForJSON(source interface{}, level int) (target interface{}) {
	refl := reflect.ValueOf(source)
	switch refl.Type().String() {
	case "map[interface {}]interface {}":
		t := map[string]interface{}{}
		fuckingSource := source.(map[interface{}]interface{})
		for key, value := range fuckingSource {
			t[fmt.Sprint(key)] = fixYamlMapsForJSON(value, level+1)
		}
		return t
	case "[]interface {}":
		sArray := source.([]interface{})
		tArray := make([]interface{}, len(sArray))
		for i, element := range sArray {
			tArray[i] = fixYamlMapsForJSON(element, level+1)
		}
		return tArray
	default:
		return source
	}
}
