package main

import (
	"bytes"
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"log"
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

var args CmdArgs

var config Config

var settings Settings

func main() {
	var r *git.Repository = &git.Repository{}
	Init()
	allRepoData := []RepoData{}
	for _, repo := range config.Repos {
		data := CloneAndGetData(repo, r)
		allRepoData = append(allRepoData, data)
		RenderLogPage(data, r)
		RenderAllCommitPages(data, r)
		RenderAllFilesPage(data)
		RenderSingleFilePages(data)
	}
	RenderIndexPage(allRepoData)
}

func Init() {
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))
	args = DefaultCmdArgs()
	settings = DefaultSettings()
	flag.StringVar(&args.ConfigFile, "config", "", "Config file.")
	flag.StringVar(&args.OutputDir, "output", "", "Dir of output.")
	flag.StringVar(&args.CloneDir, "clone", "", "Dir to clone into. Default is /tmp/${rand}")
	flag.BoolVar(&args.Silent, "silent", false, "Run in silent mode.")
	flag.Parse()

	if args.CloneDir == "" {
		args.CloneDir = fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32())
	}

	debug("config '%v'", args.ConfigFile)
	debug("output '%v'", args.OutputDir)
	debug("clone '%v'", args.CloneDir)
	configFileBytes, err := os.ReadFile(args.ConfigFile)
	configString := string(configFileBytes)
	checkErr(err)
	config = ParseConfiguration(configString)
	debug("base_url '%v'", config.BaseURL)
	debug("site_name '%v'", config.SiteName)
}

func CloneAndGetData(repo Repo, r *git.Repository) RepoData {
	err := os.MkdirAll(path.Join(args.CloneDir, repo.Name), 0755)
	checkErr(err)
	err = os.MkdirAll(path.Join(args.OutputDir, repo.Name), 0755)
	checkErr(err)
	repoRef, err := git.PlainClone(path.Join(args.CloneDir, repo.Name), false, &git.CloneOptions{
		URL: repo.URL,
	})
	checkErr(err)
	data := RepoData{
		Name:            repo.Name,
		PublishedGitURL: repo.PublishedGitURL,
		Description:     repo.Description,
		BaseURL:         config.BaseURL,
		ReadMePath:      findFileInRoot(repo.Name, settings.AllowedReadMeFiles),
		LicenseFilePath: findFileInRoot(repo.Name, settings.AllowedLicenseFiles),
	}
	*r = *repoRef
	return data
}

type LogWriter struct{}

func (writer LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func checkErr(err error) {
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}
}

func debug(format string, a ...any) {
	if !args.Silent {
		log.Printf("DEBUG: "+format, a...)
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

func findFileInRoot(name string, oneOfThese map[string]bool) string {
	dir, err := os.ReadDir(path.Join(args.CloneDir, name))
	checkErr(err)
	for _, e := range dir {
		name := e.Name()
		if _, ok := oneOfThese[name]; ok {
			return name
		}
	}
	return ""
}

type CmdArgs struct {
	Silent     bool
	ConfigFile string
	OutputDir  string
	CloneDir   string
}

func DefaultCmdArgs() CmdArgs {
	return CmdArgs{
		Silent:     true,
		ConfigFile: "",
		OutputDir:  "",
		CloneDir:   "",
	}
}

type Settings struct {
	TextExtensions      map[string]bool
	PlainFiles          map[string]bool
	AllowedLicenseFiles map[string]bool
	AllowedReadMeFiles  map[string]bool
}

func DefaultSettings() Settings {
	return Settings{
		TextExtensions: map[string]bool{
			".c":              true,
			".cc":             true,
			".conf":           true,
			".config":         true,
			".cpp":            true,
			".cs":             true,
			".css":            true,
			".csv":            true,
			".Dockerfile":     true,
			".dot":            true,
			".eslintignore":   true,
			".eslintrc":       true,
			".bashrc":         true,
			".zshrc":          true,
			".zshprofile":     true,
			".g4":             true,
			".gitignore":      true,
			".gitmodules":     true,
			".go":             true,
			".h":              true,
			".htm":            true,
			".html":           true,
			".iml":            true,
			".interp":         true,
			".java":           true,
			".js":             true,
			".json":           true,
			".jsx":            true,
			".less":           true,
			".lock":           true,
			".log":            true,
			".Makefile":       true,
			".md":             true,
			".mod":            true,
			".php":            true,
			".prettierignore": true,
			".py":             true,
			".rb":             true,
			".rs":             true,
			".scss":           true,
			".sql":            true,
			".sum":            true,
			".svg":            true,
			".tokens":         true,
			".toml":           true,
			".ts":             true,
			".tsv":            true,
			".tsx":            true,
			".txt":            true,
			".xml":            true,
			".yaml":           true,
			".yml":            true,
		},
		PlainFiles: map[string]bool{
			"Dockerfile":  true,
			"license-mit": true,
			"LICENSE-MIT": true,
			"license":     true,
			"LICENSE":     true,
			"Makefile":    true,
			"readme":      true,
			"Readme":      true,
			"ReadMe":      true,
			"README":      true,
		},
		AllowedLicenseFiles: map[string]bool{
			"license-mit": true,
			"LICENSE-MIT": true,
			"license.md":  true,
			"LICENSE.md":  true,
			"license.txt": true,
			"LICENSE.txt": true,
			"LICENSE":     true,
		},
		AllowedReadMeFiles: map[string]bool{
			"readme.md":  true,
			"Readme.md":  true,
			"ReadMe.md":  true,
			"README.md":  true,
			"readme.txt": true,
			"README.txt": true,
			"README":     true,
		},
	}
}
