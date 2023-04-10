package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/BurntSushi/toml"
)

type config struct {
	Site  siteConfig   `toml:"site"`
	Repos []repoConfig `toml:"repos"`
}

func (c *config) validate() {
	names := map[string]bool{}
	for _, r := range c.Repos {
		_, duplicate := names[r.Name]
		if duplicate {
			checkErr(errors.New(fmt.Sprintf("duplicate repo name: '%s'", r.Name)))
		}
		r.validate()
		names[r.Name] = true
	}
}

type repoConfig struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	URL         string `toml:"url"`
	AltLink     string `toml:"alt_link"`
}

func (r *repoConfig) validate() {
	ok, err := regexp.MatchString(`[A-Za-z0-9_.-]`, r.Name)
	checkErr(err)
	if !ok {
		checkErr(errors.New("repo names only allow [A-Za-z0-9_.-]"))
	}
	if r.Name == "git" {
		checkErr(errors.New("repo in config cannot have the name 'git'"))
	}
}

type siteConfig struct {
	BaseURL string `toml:"base_url"`
	Name    string `toml:"name"`
}

// cloneDir gets the directory that this repo was cloned into using the output directory
// from the program arguments, and this repo's name.
func (r *repoConfig) cloneDir() string {
	return path.Join(args.OutputDir, r.Name, "git")
}

func (r *repoConfig) findFileInRoot(oneOfThese map[string]bool) string {
	dir, err := os.ReadDir(r.cloneDir())
	checkErr(err)
	for _, e := range dir {
		name := e.Name()
		if _, ok := oneOfThese[name]; ok {
			return name
		}
	}
	return ""
}

func parseConfig(data string) config {
	conf := config{}
	_, err := toml.Decode(data, &conf)
	checkErr(err)
	return conf
}

type repoData struct {
	repoConfig
	AltLink         string
	BaseURL         string
	HeadData        HeadData
	ReadMePath      string
	LicenseFilePath string
}

type HeadData struct {
	BaseURL  string
	SiteName string
}
