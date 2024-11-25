package config

import (
	"log"
	"os"
	"runtime"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type FetchContent struct {
	Html       bool `yaml:"html" default:"true"`
	CssJsOther bool `yaml:"css_js_other" default:"false"`
	ScreenShot bool `yaml:"screenshot" default:"false"`
}

type Timeout struct {
	Navigate int `yaml:"navigate" default:"60"`
	Fetch    int `yaml:"fetch" default:"5"`
}

type Config struct {
	ThreadMax     int          `yaml:"thread_max" default:"1"`
	WaitTime      int          `yaml:"wait_time" default:"1"`
	Duplicate     string       `yaml:"duplicate" default:"same-domain"`
	FetchContents FetchContent `yaml:"fetch_contents"`
	SeedFile      string       `yaml:"seed_file" default:""`
	RandomSeed    bool         `yaml:"random_seed" default:"false"`
	OutputDir     string       `yaml:"output_dir" default:"out"`
	Timeout       Timeout      `yaml:"timeout"`
	Hops          int          `yaml:"hops" default:"-1"`
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

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(c)

	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

func LoadConf(path string) error {
	yml, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, &Configs)
	if err != nil {
		return err
	}
	if Configs.Duplicate != "same-url" && Configs.Duplicate != "same-domain" && Configs.Duplicate != "none" {
		log.Fatal("duplicate must be same-url or same-domain")
	}

	if Configs.ThreadMax < 0 {
		Configs.ThreadMax = runtime.NumCPU()
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
