package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FetchContent struct {
	Html       bool `yaml:"html"`
	CssJsOther bool `yaml:"css_js_other"`
	ScreenShot bool `yaml:"screenshot"`
}

type Config struct {
	ThreadMax     int          `yaml:"thread_max"`
	Duplicate     string       `yaml:"duplicate"`
	FetchContents FetchContent `yaml:"fetch_contents"`
	SeedFile      string       `yaml:"seed_file"`
	RandomSeed    bool         `yaml:"random_seed"`
	OutputDir     string       `yaml:"output_dir"`
}

type Env struct {
	User     string
	Password string
	DBHost   string
	DBPort   string
	DBName   string
}

var Configs *Config
var Envs *Env

func LoadConf(path string) error {
	yml, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, &Configs)
	if err != nil {
		return err
	}

	Envs = &Env{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBHost:   os.Getenv("DB_HOST"),
		DBPort:   os.Getenv("DB_PORT"),
		DBName:   os.Getenv("POSTGRES_DB"),
	}

	return nil
}
