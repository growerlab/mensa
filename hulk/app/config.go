package app

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const configPath = "conf/config.yaml"

var Conf *Config

type RedisConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	MaxIdle     int    `json:"max_idle"`
	MaxActive   int    `json:"max_active"`
	IdleTimeout int    `json:"idle_timeout"`
	Namespace   string `json:"namespace"`
}

type Config struct {
	Debug bool `json:"debug"`
	DB    struct {
		Url string `json:"url"`
	} `json:"db"`

	Redis *RedisConfig `json:"redis"`
}

func InitConfig() error {
	confRaw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.WithStack(err)
	}

	err = yaml.Unmarshal(confRaw, Conf)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
