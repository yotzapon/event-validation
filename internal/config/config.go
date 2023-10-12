package config

import (
	"event-validation/internal/repo/git"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	environment = "ENV"
)

type Config struct {
	Server struct {
		Port            int
		Env             string
		Timeout         time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Git git.Config
}

func ReadConfigFile(configModel interface{}, systemEnv string) error {
	function := "internal.config.ReadFile()"

	if systemEnv != "" {
		environment = systemEnv
	}

	env := os.Getenv(environment)

	currentDir, _ := os.Getwd()
	rViper := viper.New()
	rViper.SetConfigType("yml")
	rViper.AddConfigPath(currentDir)
	rViper.SetConfigName("config") //configs file name

	//Find and Read configs file
	if err := rViper.ReadInConfig(); err != nil {
		return fmt.Errorf("%v: cannot read configuration file, %v", function, err)
	}

	var allEnvConfig map[string]interface{}
	if err := rViper.Unmarshal(&allEnvConfig); err != nil {
		return fmt.Errorf("%v: cannot unmarshal configuration, %v", function, err)
	}

	// get configuration by environment(local, dev, or etc.) and marshal to binary
	envConfig, err := yaml.Marshal(allEnvConfig[env])
	if err != nil {
		return fmt.Errorf("%s: unable to marshal configuration data at env=%s, %s", function, env, err.Error())
	}

	err = yaml.Unmarshal(envConfig, configModel)
	if err != nil {
		return fmt.Errorf("%s: unable to unmarshal configuration data to Config model, %s", function, err.Error())
	}
	return nil
}
