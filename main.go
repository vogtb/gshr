package main

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
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
	RepoData       RepoData
	TextExtensions map[string]bool
	PlainFiles     map[string]bool
}

func DefaultConfig() Config {
	return Config{
		DebugOn:   true,
		Repo:      "",
		OutputDir: "",
		CloneDir:  "",
		RepoData: RepoData{
			Name:        "",
			Description: "",
			BaseURL:     "/",
		},
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

type RepoData struct {
	Name            string
	GitURL          string
	Description     string
	BaseURL         string
	HasReadMe       bool
	ReadMePath      string
	HasLicenseFile  bool
	LicenseFilePath string
}

func main() {
	config = DefaultConfig()
	flag.StringVar(&config.Repo, "repo", "", "Repo to use.")
	flag.BoolVar(&config.DebugOn, "debug", true, "Run in debug mode.")
	flag.StringVar(&config.OutputDir, "output", "", "Dir of output.")
	flag.StringVar(&config.CloneDir, "clone", "", "Directory to clone into. Defaults to /tmp/${rand}")
	flag.StringVar(&config.RepoData.BaseURL, "base-url", "/", "Base URL for serving.")
	flag.StringVar(&config.RepoData.GitURL, "git-url", "", "Show where repo is hosted.")
	flag.StringVar(&config.RepoData.Name, "name", "untitled repo", "Name for display")
	flag.StringVar(&config.RepoData.Description, "desc", "untitled repo", "Description for display")
	flag.Parse()

	if config.Repo == "" {
		checkErr(errors.New("--repo flag is required"))
	}

	if config.CloneDir == "" {
		config.CloneDir = fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32())
	}

	config.RepoData.BaseURL = path.Join(config.RepoData.BaseURL, "/")

	debug("repo = %v", config.Repo)
	debug("output = %v", config.OutputDir)
	debug("clone = %v", config.CloneDir)
	debug("base-url = %v", config.BaseURL)
	r := CloneAndInfo()
	config.RepoData.ReadMePath = getReadmePath()
	config.RepoData.HasReadMe = config.RepoData.ReadMePath != ""
	config.RepoData.LicenseFilePath = getLicenseFilePath()
	config.RepoData.HasLicenseFile = config.RepoData.LicenseFilePath != ""
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

func syntaxHighlightTools(pathOrExtension string) (chroma.Lexer, *chroma.Style, *html.Formatter) {
	lexer := lexers.Match(pathOrExtension)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("borland")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(
		html.WithClasses(true),
		html.WithLineNumbers(true),
		html.LinkableLineNumbers(true, ""),
	)
	return lexer, style, formatter
}

func highlight(pathOrExtension string, data *string) string {
	lexer, style, formatter := syntaxHighlightTools(pathOrExtension)
	iterator, err := lexer.Tokenise(nil, *data)
	buf := bytes.NewBufferString("")
	err = formatter.Format(buf, style, iterator)
	checkErr(err)
	return buf.String()
}

func getReadmePath() string {
	for _, file := range []string{
		"readme.md",
		"README.md",
		"readme.txt",
		"README.txt",
		"README",
	} {
		if stat, err := os.Stat(path.Join(config.CloneDir, file)); err == nil {
			return stat.Name()
		}
	}
	return ""
}

func getLicenseFilePath() string {
	for _, file := range []string{
		"license-mit",
		"LICENSE-MIT",
		"license.md",
		"LICENSE.md",
		"license.txt",
		"LICENSE.txt",
		"LICENSE",
	} {
		if stat, err := os.Stat(path.Join(config.CloneDir, file)); err == nil {
			return stat.Name()
		}
	}
	return ""
}
