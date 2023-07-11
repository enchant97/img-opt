package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func (c Config) LoadFromYaml(fp string) (Config, error) {
	rawConfig, err := os.ReadFile(fp)
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		return c, err
	}
	validate := validator.New()
	err = validate.Struct(c)
	return c, err
}
