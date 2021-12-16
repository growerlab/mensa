package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/growerlab/backend/app/utils/conf"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const DefaultConfigPath = "conf/config.yaml"
const DefaultENV = "dev"

var envConfig map[string]*Config
var config *Config

type Redis struct {
	conf.Redis          `yaml:",inline"`
	PermissionNamespace string `yaml:"permission_namespace"`
}

type Config struct {
	Debug         bool   `yaml:"-"`
	User          string `yaml:"user"`
	Listen        string `yaml:"listen"`
	HttpListen    string `yaml:"http_listen"`
	GitPath       string `yaml:"git_path"`
	Deadline      int64  `yaml:"deadline"`
	IdleTimeout   int64  `yaml:"idle_timeout"`
	GitRepoDir    string `yaml:"git_repo_dir"`
	DBUrl         string `yaml:"db_url"`
	Redis         *Redis `yaml:"redis"`
	GoGitGrpcAddr string `yaml:"go_git_grpc_addr"`
}

func (c *Config) validate() error {
	if c.User == "" {
		return errors.New("git uesr is required")
	}
	if !strings.Contains(c.Listen, ":") || !strings.Contains(c.HttpListen, ":") {
		return errors.New("listen addr is invalid")
	}

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

	err = yaml.UnmarshalStrict(rawConfig, &envConfig)
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

	if env == DefaultENV {
		// for dev
		config.Debug = true
		config.GitRepoDir, err = filepath.Abs(config.GitRepoDir)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return config.validate()
}

func GetConfig() *Config {
	return config
}
