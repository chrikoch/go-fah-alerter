package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//Config is a whole config
type Config struct {
	UserNames []string `json:"usernames"`
}

//ReadFromFile reads file filename into Config struct
func (c *Config) ReadFromFile(filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &c)

	if err != nil {
		return err
	}

	return nil
}
