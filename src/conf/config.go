package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const DefaultConfigPath = "conf/config.yaml"
const DefaultENV = "dev"

var envConfig map[string]*Config
var config *Config

type Config struct {
	User       string `yaml:"user"`
	Listen     string `yaml:"listen"`
	HttpListen string `yaml:"http_listen"`
	// HostKeys    []string `yaml:"host_keys"`
	GitPath     string `yaml:"git_path"`
	Deadline    int    `yaml:"deadline"`
	IdleTimeout int    `yaml:"idle_timeout"`
	GitRepoDir  string `yaml:"git_repo_dir"`
}

func (c *Config) validate() error {
	if c.User == "" {
		return errors.New("git uesr is required")
	}
	if !strings.Contains(c.Listen, ":") || !strings.Contains(c.HttpListen, ":") {
		return errors.New("listen addr is invalid")
	}
	// for _, k := range c.HostKeys {
	// 	if _, err := os.Stat(k); os.IsNotExist(err) {
	// 		return errors.WithMessage(err, "host keys")
	// 	}
	// }
	if _, err := os.Stat(c.GitPath); os.IsNotExist(err) {
		return errors.WithMessage(err, "git path")
	}
	if _, err := os.Stat(c.GitRepoDir); os.IsNotExist(err) {
		return errors.WithMessage(err, "git repo dir")
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
	env := os.Getenv("ENV")
	if env == "" {
		env = DefaultENV
	}
	config, ok = envConfig[env]
	if !ok {
		return errors.New("not found config by env: " + env)
	}
	// for dev
	if !strings.HasPrefix(config.GitRepoDir, "/") && env == "dev" {
		config.GitRepoDir = filepath.Join(os.Getenv("GOPATH"), "src", "github.com/growerlab/mensa", config.GitRepoDir)
	}
	return config.validate()
}

func GetConfig() *Config {
	return config
}
