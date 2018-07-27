package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"gopx.io/gopx-common/log"
)

// ServiceConfigPath holds API service related configuration file path.
const ServiceConfigPath = "./config/service.json"

// ServiceConfig represents API service related configurations.
type ServiceConfig struct {
	Host         string        `json:"host"`
	UseHTTP      bool          `json:"useHTTP"`
	HTTPPort     int           `json:"HTTPPort"`
	UseHTTPS     bool          `json:"useHTTPS"`
	HTTPSPort    int           `json:"HTTPSPort"`
	CertFile     string        `json:"certFile"`
	KeyFile      string        `json:"keyFile"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
	IdleTimeout  time.Duration `json:"idleTimeout"`
}

// Service holds loaded API service related configurations.
var Service = new(ServiceConfig)

func init() {
	bytes, err := ioutil.ReadFile(ServiceConfigPath)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
	err = json.Unmarshal(bytes, Service)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
}
