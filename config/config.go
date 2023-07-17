package config

type TypeFormatConfig struct {
	Quality int `yaml:"quality" validate:"required"`
}

type TypeConfig struct {
	MaxWidth int                         `yaml:"max_width" validate:"required"`
	Formats  map[string]TypeFormatConfig `yaml:"formats" validate:"required"`
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
	Enable   bool `yaml:"enable" validate:"required"`
	MaxWidth *int `yaml:"max_width"`
	AVIF     bool `yaml:"avif"`
}

type BrowserTTLConfig struct {
	Max      uint `yaml:"max" validate:"required"`
	MaxStale uint `yaml:"max_stale" validate:"required"`
}

type Config struct {
	Bind          BindConfig         `yaml:"bind" validate:"required"`
	Metrics       bool               `yaml:"metrics"`
	JobLimit      uint               `yaml:"job_limit"`
	BrowserTTL    *BrowserTTLConfig  `yaml:"browser_ttl"`
	OriginalsBase string             `yaml:"originals_base" validate:"required"`
	AutoOptimize  AutoOptimizeConfig `yaml:"auto_optimize" validate:"required"`
	TypeOptimize  TypeOptimizeConfig `yaml:"type_optimize" validate:"required"`
}
