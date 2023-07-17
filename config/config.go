package config

type FormatConfig struct {
	Quality int `yaml:"quality" validate:"required"`
}

type PresetConfig struct {
	MaxWidth int                     `yaml:"max_width" validate:"required"`
	Formats  map[string]FormatConfig `yaml:"formats" validate:"required"`
}

type PresetOptimizeConfig struct {
	Enable  bool                    `yaml:"enable" validate:"required"`
	Presets map[string]PresetConfig `yaml:"presets"`
}

type TLSConfig struct {
	CertFile string `yaml:"cert_file" validate:"required"`
	KeyFile  string `yaml:"key_file" validate:"required"`
}

type BindConfig struct {
	Host string     `yaml:"host" validate:"required"`
	Port uint       `yaml:"port" validate:"required"`
	TLS  *TLSConfig `yaml:"tls"`
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
	Bind           BindConfig           `yaml:"bind" validate:"required"`
	Metrics        bool                 `yaml:"metrics"`
	JobLimit       uint                 `yaml:"job_limit"`
	BrowserTTL     *BrowserTTLConfig    `yaml:"browser_ttl"`
	OriginalsBase  string               `yaml:"originals_base" validate:"required"`
	AutoOptimize   AutoOptimizeConfig   `yaml:"auto_optimize" validate:"required"`
	PresetOptimize PresetOptimizeConfig `yaml:"preset_optimize" validate:"required"`
}
