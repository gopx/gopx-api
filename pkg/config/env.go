package config

import (
	"encoding/json"
	"io/ioutil"

	"gopx.io/gopx-common/log"
)

// EnvConfigPath holds the environment variables file path.
const EnvConfigPath = "./config/env.json"

// EnvConfig represents the mapping of environment variables.
type EnvConfig struct {
	GoPxVCSAPIIP      string `json:"GoPxVCSAPIIP"`
	GoPxVCSAPIPort    string `json:"GoPxVCSAPIPort"`
	GoPxVCSAPIAuthKey string `json:"GoPxVCSAPIAuthKey"`
}

// Env holds the mapping of environment variables.
var Env = new(EnvConfig)

func init() {
	bytes, err := ioutil.ReadFile(EnvConfigPath)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
	err = json.Unmarshal(bytes, Env)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
}
