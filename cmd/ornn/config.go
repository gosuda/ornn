package ornn

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	DB  ConfigDB
	Gen ConfigGen
}

type ConfigDB struct {
	Type     string `mapstructure:"Type"`
	Path     string `mapstructure:"Path"`
	Addr     string `mapstructure:"Addr"`
	Port     string `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	Name     string `mapstructure:"Name"`
}

type ConfigGen struct {
	SchemaPath  string `mapstructure:"SchemaPath"`
	ConfigPath  string `mapstructure:"ConfigPath"`
	GenPath     string `mapstructure:"GenPath"`
	FileName    string `mapstructure:"FileName"`
	PackageName string `mapstructure:"PackageName"`
	ClassName   string `mapstructure:"ClassName"`
}

func loadConfig() (*Config, error) {
	var k = koanf.New(".")
	if configFilePath == "" {
		configFilePath = "config.toml"
	}

	if err := k.Load(file.Provider(configFilePath), toml.Parser()); err != nil {
		return nil, err
	}
	var config Config
	if err := k.Unmarshal("DB", &config.DB); err != nil {
		return nil, err
	}

	if err := k.Unmarshal("Gen", &config.DB); err != nil {
		return nil, err
	}

	return &config, nil
}
