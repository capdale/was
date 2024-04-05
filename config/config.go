package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrInvalidConfig = errors.New("invalid config file")
var ErrInvalidCredForm = errors.New("invalid cred form")

type Config struct {
	Service Service `yaml:"service"`
	Mysql   Mysql   `yaml:"mysql"`
	Redis   Redis   `yaml:"redis"`
	Rpc     Rpc     `yaml:"rpc"`
	Key     Key     `yaml:"key"`
	Oauth   Oauth   `yaml:"oauth"`
	Storage Storage `yaml:"storage"`
	Email   Email   `yaml:"email"`
}

type Service struct {
	Address string `yaml:"address"`
	Cors    Cors   `yaml:"cors"`
	Log     Log    `yaml:"log"`
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

type Log struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Console    bool   `yaml:"console"`
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
	ImageClassify RpcObject `yaml:"imageClassify"`
}

type RpcObject struct {
	Address []string `yaml:"address"`
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

type Storage struct {
	S3    *S3    `yaml:"s3,omitempty"`
	Local *Local `yaml:"local,omitempty"`
}

type Local struct {
	BaseDir string `yaml:"baseDirectory"`
}

type S3 struct {
	Name   string  `yaml:"bucketName"`
	Region string  `yaml:"region"`
	Id     *string `yaml:"id,omitempty"`
	Key    *string `yaml:"key,omitempty"`
}

type Email struct {
	Ses *Ses `yaml:"ses,omitempty"`
}

type Ses struct {
	Region string  `yaml:"region"`
	Domain string  `yaml:"domain"`
	Id     *string `yaml:"id,omitempty"`
	Key    *string `yaml:"key,omitempty"`
}

func ParseConfig(filepath string) (c *Config, err error) {
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	c = &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return
	}

	if c.Storage.Local == nil && c.Storage.S3 == nil {
		err = ErrInvalidConfig
		return
	}
	if c.Email.Ses == nil {
		err = ErrInvalidConfig
		return
	}
	return
}
