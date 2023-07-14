package config

type TypeFormatConfig struct {
	Quality int `yaml:"quality" validate:"required"`
}

type TypeConfig struct {
	Width   int                         `yaml:"width" validate:"required"`
	Formats map[string]TypeFormatConfig `yaml:"formats" validate:"required"`
}

type TypeOptimizeConfig struct {
	Enable bool                  `yaml:"enable" validate:"required"`
	Types  map[string]TypeConfig `yaml:"types"`
}

type BindConfig struct {
	Host string `yaml:"host" validate:"required"`
	Port uint   `yaml:"port" validate:"required"`
}

type AutoOptimizeConfig struct {
	Enable bool `yaml:"enable" validate:"required"`
}

type Config struct {
	Bind          BindConfig         `yaml:"bind" validate:"required"`
	OriginalsBase string             `yaml:"originals_base" validate:"required"`
	AutoOptimize  AutoOptimizeConfig `yaml:"auto_optimize" validate:"required"`
	TypeOptimize  TypeOptimizeConfig `yaml:"type_optimize" validate:"required"`
}
