package config

type OptimizationFormatConfig struct {
	Enable  bool `yaml:"enable" validate:"required"`
	Quality uint `yaml:"quality" validate:"required"`
}

type OptimizationFormatsConfig struct {
	JPEG OptimizationFormatConfig `yaml:"jpeg"`
	WebP OptimizationFormatConfig `yaml:"webp"`
	AVIF OptimizationFormatConfig `yaml:"avif"`
}

type OptimizationConfig struct {
	Name    string                    `yaml:"name" validate:"required"`
	Width   uint                      `yaml:"width" validate:"required"`
	Formats OptimizationFormatsConfig `yaml:"formats" validate:"required"`
}

type DefinedConfig struct {
	Enable        bool                 `yaml:"enable" validate:"required"`
	Optimizations []OptimizationConfig `yaml:"optimizations"`
}

type BindConfig struct {
	Host string `yaml:"host" validate:"required"`
	Port uint   `yaml:"port" validate:"required"`
}

type Config struct {
	Bind          BindConfig    `yaml:"bind" validate:"required"`
	OriginalsBase string        `yaml:"originals_base" validate:"required"`
	Defined       DefinedConfig `yaml:"defined"`
}
