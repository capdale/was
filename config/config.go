package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service Service `yaml:"service"`
	Mysql   Mysql   `yaml:"mysql"`
	Redis   Redis   `yaml:"redis"`
	Rpc     Rpc     `yaml:"rpc"`
	Key     Key     `yaml:"key"`
	Oauth   Oauth   `yaml:"oauth"`
}

type Service struct {
	Address string `yaml:"address"`
	Cors    Cors   `yaml:"cors"`
}

type Cors struct {
	AllowOrigins           []string `yaml:"allowOrigins"`
	AllowMethods           []string `yaml:"allowMethods"`
	AllowHeaders           []string `yaml:"allowHeaders"`
	AllowCredentials       bool     `default:"true" yaml:"allowCredentials"`
	ExposeHeaders          []string `yaml:"exposeHeaders"`
	MaxAge                 int      `default:"3600" yaml:"maxAge"`
	AllowWildcard          bool     `default:"false" yaml:"allowWildcard"`
	AllowBrowserExtensions bool     `default:"false" yaml:"allowBrowserExtensions"`
	AllowWebSockets        bool     `default:"true" yaml:"allowWebSockets"`
	AllowFiles             bool     `default:"false" yaml:"allowFiles"`
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

type Rpc struct {
	Address string `yaml:"address"`
}

type Key struct {
	Jwtkey          string `yaml:"jwtkey"`
	SessionStateKey string `yaml:"sessionStateKey"`
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
