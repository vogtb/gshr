package main

import (
	"embed"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
)

//go:embed templates/*
var htmlTemplates embed.FS

var config Config

type Config struct {
	DebugOn        bool
	Repo           string
	OutputDir      string
	CloneDir       string
	BaseURL        string
	TextExtensions map[string]bool
	PlainFiles     map[string]bool
}

func DefaultConfig() Config {
	return Config{
		DebugOn:   true,
		Repo:      "",
		OutputDir: "",
		BaseURL:   "/",
		CloneDir:  "",
		TextExtensions: map[string]bool{
			".c":          true,
			".cc":         true,
			".conf":       true,
			".config":     true,
			".cpp":        true,
			".cs":         true,
			".css":        true,
			".csv":        true,
			".Dockerfile": true,
			".gitignore":  true,
			".gitmodules": true,
			".go":         true,
			".h":          true,
			".htm":        true,
			".html":       true,
			".iml":        true,
			".js":         true,
			".json":       true,
			".jsx":        true,
			".less":       true,
			".lock":       true,
			".log":        true,
			".Makefile":   true,
			".md":         true,
			".mod":        true,
			".php":        true,
			".py":         true,
			".rb":         true,
			".rs":         true,
			".scss":       true,
			".sql":        true,
			".sum":        true,
			".svg":        true,
			".toml":       true,
			".ts":         true,
			".tsv":        true,
			".tsx":        true,
			".txt":        true,
			".xml":        true,
			".yaml":       true,
			".yml":        true,
		},
		PlainFiles: map[string]bool{
			"Dockerfile": true,
			"LICENSE":    true,
			"Makefile":   true,
			"readme":     true,
			"README":     true,
		},
	}
}

func main() {
	config = DefaultConfig()
	flag.StringVar(&config.Repo, "repo", "", "Repo to use.")
	flag.BoolVar(&config.DebugOn, "debug", true, "Run in debug mode.")
	flag.StringVar(&config.OutputDir, "output", "", "Dir of output.")
	flag.StringVar(&config.CloneDir, "clone", "", "Directory to clone into. Defaults to /tmp/${rand}")
	flag.StringVar(&config.BaseURL, "base-url", "/", "Base URL for loading styles.")
	flag.Parse()

	if config.Repo == "" {
		checkErr(errors.New("--repo flag is required"))
	}

	if config.CloneDir == "" {
		config.CloneDir = fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32())
	}

	config.BaseURL = path.Join(config.BaseURL, "/")

	debug("repo = %v", config.Repo)
	debug("output = %v", config.OutputDir)
	debug("clone = %v", config.CloneDir)
	debug("base-url = %v", config.BaseURL)
	r := CloneAndInfo()
	RenderLogPage(r)
	RenderAllCommitPages(r)
	RenderAllFilesPage()
	RenderSingleFilePages()
}

func CloneAndInfo() *git.Repository {
	r, err := git.PlainClone(config.CloneDir, false, &git.CloneOptions{
		URL: config.Repo,
	})
	checkErr(err)
	return r
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func debug(format string, a ...any) {
	if config.DebugOn {
		fmt.Printf(format, a...)
		fmt.Print("\n")
	}
}
