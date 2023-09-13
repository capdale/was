package server

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mysql Mysql `yaml:"mysql"`
	Redis Redis `yaml:"redis"`
	Key   Key   `yaml:"key"`
}

type Mysql struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type Key struct {
	Jwtkey string `yaml:"jwtkey"`
}

func parseConfig(filepath string) (c *Config, err error) {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	c = &Config{}
	err = yaml.Unmarshal(buf, c)
	return

}
