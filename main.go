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

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-git/go-git/v5"
)

//go:embed templates/*
var htmlTemplates embed.FS

var config Config

func main() {
	var r *git.Repository = &git.Repository{}
	Init()
	CloneAndInfo(r)
	RenderLogPage(r)
	RenderAllCommitPages(r)
	RenderAllFilesPage()
	RenderSingleFilePages()
}

func Init() {
	config = DefaultConfig()
	flag.StringVar(&config.Repo, "repo", "", "Repo to use.")
	flag.BoolVar(&config.DebugOn, "debug", true, "Run in debug mode.")
	flag.StringVar(&config.OutputDir, "output", "", "Dir of output.")
	flag.StringVar(&config.CloneDir, "clone", "", "Dir to clone into. Default is /tmp/${rand}")
	flag.StringVar(&config.RepoData.BaseURL, "base-url", "/", "Base URL for serving.")
	flag.StringVar(&config.RepoData.GitURL, "git-url", "", "Show where repo is hosted.")
	flag.StringVar(&config.RepoData.Description, "desc", "<no description>", "Description to show.")
	flag.Parse()

	if config.Repo == "" {
		checkErr(errors.New("--repo flag is required"))
	}

	if config.CloneDir == "" {
		config.CloneDir = fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32())
	}

	config.RepoData.BaseURL = path.Join(config.RepoData.BaseURL, "/")
	config.RepoData.Name = path.Clean(path.Base(config.Repo))

	debug("repo = %v", config.Repo)
	debug("output = %v", config.OutputDir)
	debug("clone = %v", config.CloneDir)
	debug("base-url = %v", config.RepoData.BaseURL)
}

func CloneAndInfo(r *git.Repository) {
	repo, err := git.PlainClone(config.CloneDir, false, &git.CloneOptions{
		URL: config.Repo,
	})
	checkErr(err)
	config.RepoData.ReadMePath = findFileInRoot(config.AllowedReadMeFiles)
	config.RepoData.LicenseFilePath = findFileInRoot(config.AllowedLicenseFiles)
	*r = *repo
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

func highlight(pathOrExtension string, data *string) string {
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
	iterator, err := lexer.Tokenise(nil, *data)
	buf := bytes.NewBufferString("")
	err = formatter.Format(buf, style, iterator)
	checkErr(err)
	return buf.String()
}

func findFileInRoot(oneOfThese map[string]bool) string {
	dir, err := os.ReadDir(config.CloneDir)
	checkErr(err)
	for _, e := range dir {
		name := e.Name()
		if _, ok := oneOfThese[name]; ok {
			return name
		}
	}
	return ""
}
