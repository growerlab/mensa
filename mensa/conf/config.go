package conf

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const DefaultConfigPath = "conf/config.yaml"
const DefaultENV = "env"

var envConfig map[string]*Config
var config *Config

type Config struct {
	User        string   `yaml:"user"`
	Listen      string   `yaml:"listen"`
	HttpListen  string   `yaml:"http_listen"`
	HostKeys    []string `yaml:"host_keys"`
	Deadline    int      `yaml:"deadline"`
	IdleTimeout int      `yaml:"idle_timeout"`
}

func LoadConfig() error {
	var ok bool
	rawConfig, err := ioutil.ReadFile(DefaultConfigPath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = yaml.Unmarshal(rawConfig, &envConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	env := os.Getenv(DefaultENV)
	config, ok = envConfig[env]
	if !ok {
		return errors.New("not found config by env: " + env)
	}
	return nil
}

func GetConfig() *Config {
	return config
}
