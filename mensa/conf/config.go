package conf

import (
	"io/ioutil"
	"os"
	"strings"

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
	GitPath     string   `yaml:"git_path"`
	Deadline    int      `yaml:"deadline"`
	IdleTimeout int      `yaml:"idle_timeout"`
}

func (c *Config) validate() error {
	if c.User == "" {
		return errors.New("git uesr is required")
	}
	if !strings.Contains(c.Listen, ":") || !strings.Contains(c.HttpListen, ":") {
		return errors.New("listen addr is invalid")
	}
	for _, k := range c.HostKeys {
		if _, err := os.Stat(k); os.IsNotExist(err) {
			return errors.WithStack(err)
		}
	}
	if _, err := os.Stat(c.GitPath); os.IsNotExist(err) {
		return errors.WithStack(err)
	}
	return nil
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
	if env == "" {
		env = DefaultENV
	}
	config, ok = envConfig[env]
	if !ok {
		return errors.New("not found config by env: " + env)
	}
	return config.validate()
}

func GetConfig() *Config {
	return config
}
