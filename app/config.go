package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FetchContent struct {
    Html bool `yaml:"html"`
    CssJsOther bool `yaml:"css_js_other"`
    ScreenShot bool `yaml:"screenshot"`
}

type Config struct {
    ThreadMax int `yaml:"thread_max"`
    Duplicate string `yaml:"duplicate"`
    FetchContents FetchContent `yaml:"fetch_contents"`
    SeedFile string `yaml:"seed_file"`
    OutputDir string `yaml:"output_dir"`
}

var Configs *Config

func loadConf(path string) error {
	yml, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yml, &Configs)
	if err != nil {
		return err
	}

	return nil
}