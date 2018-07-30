package config

import (
	"encoding/json"
	"io/ioutil"

	"gopx.io/gopx-common/log"
)

// DbConfigPath holds database related configuration file path.
const DbConfigPath = "./config/db.json"

// DbConfig represents database related configurations.
type DbConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
}

// Db holds loaded database related configurations.
var Db = new(DbConfig)

func init() {
	bytes, err := ioutil.ReadFile(DbConfigPath)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
	err = json.Unmarshal(bytes, Db)
	if err != nil {
		log.Fatal("Error: %s", err)
	}
}
