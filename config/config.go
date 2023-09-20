package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mysql Mysql `yaml:"mysql"`
	Redis Redis `yaml:"redis"`
	Key   Key   `yaml:"key"`
	Oauth Oauth `yaml:"oauth"`
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

type Oauth struct {
	Github Github `yaml:"github"`
}

type Github struct {
	Id       string `yaml:"id"`
	Secret   string `yaml:"secret"`
	Redirect string `yaml:"redirect"`
}

func ParseConfig(filepath string) (c *Config, err error) {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	c = &Config{}
	err = yaml.Unmarshal(buf, c)
	return
}
