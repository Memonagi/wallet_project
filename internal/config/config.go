package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/davecgh/go-spew/spew"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

const envFileName = ".env"

type Config struct {
	env *EnvSetting
}

type EnvSetting struct {
	AppPort     int    `env:"APP_PORT" env-default:"8080" env-description:"Application port"`
	KafkaPort   string `env:"KAFKA_PORT" env-default:"localhost:9094" env-description:"Kafka port"`
	PostgresDSN string `env:"POSTGRES_DSN" env-default:"postgresql://user:password@localhost:5432/mydatabase" env-description:"PostgreSQL DSN"` //nolint:lll
}

func findConfigFile() bool {
	_, err := os.Stat(envFileName)

	return err == nil
}

func (e *EnvSetting) GetHelpString() (string, error) {
	baseHeader := "options which can be set with env: "

	helpString, err := cleanenv.GetDescription(e, &baseHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get help string: %w", err)
	}

	return helpString, nil
}

func New() *Config {
	envSetting := &EnvSetting{}

	helpString, err := envSetting.GetHelpString()
	if err != nil {
		logrus.Panicf("failed to get help string: %v", err)
	}

	logrus.Info(helpString)

	if findConfigFile() {
		if err := cleanenv.ReadConfig(envFileName, envSetting); err != nil {
			logrus.Panicf("failed to read env config: %v", err)
		}
	} else if err := cleanenv.ReadEnv(envSetting); err != nil {
		logrus.Panicf("failed to read env config: %v", err)
	}

	return &Config{env: envSetting}
}

func (c *Config) PrintDebug() {
	envReflect := reflect.Indirect(reflect.ValueOf(c.env))
	envReflectType := envReflect.Type()

	exp := regexp.MustCompile("([Tt]oken|[Pp]assword)")

	for i := range envReflect.NumField() {
		key := envReflectType.Field(i).Name

		if exp.MatchString(key) {
			val, _ := envReflect.Field(i).Interface().(string)
			logrus.Debugf("%s: len %d", key, len(val))

			continue
		}

		logrus.Debugf("%s: %v", key, spew.Sprintf("%#v", envReflect.Field(i).Interface()))
	}
}

func (c *Config) GetAppPort() int {
	return c.env.AppPort
}

func (c *Config) GetKafkaPort() string {
	return c.env.KafkaPort
}

func (c *Config) GetPostgresDSN() string {
	return c.env.PostgresDSN
}
