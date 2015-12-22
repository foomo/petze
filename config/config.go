package config

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
	Insecure bool
}

// Config a config
type Config struct {
	Services []Service
}
