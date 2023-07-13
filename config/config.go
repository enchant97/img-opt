package config

type TypeFormatConfig struct {
	Enable  bool `yaml:"enable" validate:"required"`
	Quality int  `yaml:"quality" validate:"required"`
}

type TypeConfig struct {
	Width   int                         `yaml:"width" validate:"required"`
	Formats map[string]TypeFormatConfig `yaml:"formats" validate:"required"`
}

type OptimizationConfig struct {
	Enable bool                  `yaml:"enable" validate:"required"`
	Types  map[string]TypeConfig `yaml:"Types"`
}

type BindConfig struct {
	Host string `yaml:"host" validate:"required"`
	Port uint   `yaml:"port" validate:"required"`
}

type Config struct {
	Bind          BindConfig         `yaml:"bind" validate:"required"`
	OriginalsBase string             `yaml:"originals_base" validate:"required"`
	AutoOptimize  bool               `yaml:"auto_optimize" validate:"required"`
	Optimizations OptimizationConfig `yaml:"optimizations" validate:"required"`
}
