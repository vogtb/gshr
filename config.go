package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Repos   []Repo
	BaseURL string `toml:"base_url"`
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

type CmdArgs struct {
	DebugOn    bool
	ConfigFile string
	OutputDir  string
	CloneDir   string
}

func DefaultCmdArgs() CmdArgs {
	return CmdArgs{
		DebugOn:    true,
		ConfigFile: "",
		OutputDir:  "",
		CloneDir:   "",
	}
}

type RepoData struct {
	Name            string
	GitURL          string
	Description     string
	BaseURL         string
	ReadMePath      string
	LicenseFilePath string
}
