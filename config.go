package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	SiteName string `toml:"site_name"`
	Repos    []Repo
	BaseURL  string `toml:"base_url"`
}

type Repo struct {
	Name        string
	Description string
	Path        string
	GitURL      string `toml:"git_url"`
}

func ParseConfiguration(data string) Config {
	conf := Config{}
	_, err := toml.Decode(data, &conf)
	checkErr(err)
	return conf
}

type RepoData struct {
	Name            string
	GitURL          string
	Description     string
	BaseURL         string
	ReadMePath      string
	LicenseFilePath string
}
