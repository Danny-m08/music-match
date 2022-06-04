package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var gCFG *Config

//CreateConfigFromFile creates a config object from the specified file path and sets the global config object to it
func CreateConfigFromFile(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = CreateConfigFromData(data)
	if err != nil {
		return err
	}

	return nil
}

//CreateConfigFromData creates a config object from the provided data and sets the global config object to it
func CreateConfigFromData(data []byte) error {
	config := &Config{}

	err := yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	gCFG = config
	return nil
}

//GetGlobalConfig makes the global config object retrievable from external packages
func GetGlobalConfig() *Config {
	return gCFG
}

//GetDBConfig returns the DBConfig of the global config object
func (config *Config) GetDBConfig() Neo4jConfig {
	return config.DB
}

//GetHTTPServerConfig returns the HTTP Config object from the global config
func (config *Config) GetHTTPServerConfig() HTTPConfig {
	return config.Http
}
