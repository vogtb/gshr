package main

import (
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseURL  string `toml:"base_url"`
	SiteName string `toml:"site_name"`
	Repos    []Repo `toml:"repos"`
}

type Repo struct {
	Name            string `toml:"name"`
	Description     string `toml:"description"`
	URL             string `toml:"url"`
	HostGit         bool   `toml:"host_git"`
	PublishedGitURL string `toml:"published_git_url"`
}

// / CloneDir gets the directory that this repo was cloned into using the output directory
// / from the program arguments, and this repo's name.
func (r *Repo) CloneDir() string {
	return path.Join(args.OutputDir, r.Name, "git")
}

func (r *Repo) FindFileInRoot(oneOfThese map[string]bool) string {
	dir, err := os.ReadDir(r.CloneDir())
	checkErr(err)
	for _, e := range dir {
		name := e.Name()
		if _, ok := oneOfThese[name]; ok {
			return name
		}
	}
	return ""
}

func ParseConfiguration(data string) Config {
	conf := Config{}
	_, err := toml.Decode(data, &conf)
	checkErr(err)
	return conf
}

type RepoData struct {
	Repo
	PublishedGitURL string
	BaseURL         string
	ReadMePath      string
	LicenseFilePath string
}
