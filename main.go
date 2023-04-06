package main

import (
	"bytes"
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-git/go-git/v5"
)

//go:embed template.*.html
var htmlTemplates embed.FS

//go:embed gshr.css
var css []byte

//go:embed favicon.ico
var favicon []byte

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
	RenderAssets()
	for _, repo := range config.Repos {
		HostRepo(repo)
	}
}

func Init() {
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))
	args = DefaultCmdArgs()
	settings = DefaultSettings()
	pwd, err := os.Getwd()
	checkErr(err)
	args.Wd = pwd
	flag.StringVar(&args.ConfigPath, "c", "", "Config file.")
	flag.StringVar(&args.OutputDir, "o", "", "Dir of output.")
	flag.BoolVar(&args.Silent, "s", false, "Run in silent mode.")
	flag.Parse()
	debug("working dir '%v'", args.Wd)

	if !strings.HasPrefix(args.ConfigPath, "/") {
		args.ConfigPath = path.Join(args.Wd, args.ConfigPath)
		checkFile(args.ConfigPath)
	}

	if !strings.HasPrefix(args.OutputDir, "/") {
		args.OutputDir = path.Join(args.Wd, args.OutputDir)
		checkDir(args.OutputDir)
	}

	debug("config '%v'", args.ConfigPath)
	debug("output '%v'", args.OutputDir)
	configFileBytes, err := os.ReadFile(args.ConfigPath)
	configString := string(configFileBytes)
	checkErr(err)
	config = ParseConfiguration(configString)
	debug("base_url '%v'", config.BaseURL)
	debug("site_name '%v'", config.SiteName)
}

func CloneAndGetData(repo Repo, r *git.Repository) RepoData {
	err := os.MkdirAll(repo.CloneDir(), 0755)
	checkErr(err)
	err = os.MkdirAll(path.Join(args.OutputDir, repo.Name), 0755)
	checkErr(err)
	debug("cloning '%v'", repo.Name)
	repoRef, err := git.PlainClone(repo.CloneDir(), false, &git.CloneOptions{
		URL: repo.URL,
	})
	checkErr(err)
	data := RepoData{
		Repo:            repo,
		PublishedGitURL: repo.PublishedGitURL,
		BaseURL:         config.BaseURL,
		HeadData: HeadData{
			BaseURL:  config.BaseURL,
			SiteName: config.SiteName,
		},
		ReadMePath:      repo.FindFileInRoot(settings.AllowedReadMeFiles),
		LicenseFilePath: repo.FindFileInRoot(settings.AllowedLicenseFiles),
	}
	*r = *repoRef
	return data
}

func RenderAssets() {
	debug("rendering gshr.css")
	debug("rendering favicon.ico")
	checkErr(os.WriteFile(path.Join(args.OutputDir, "gshr.css"), css, 0666))
	checkErr(os.WriteFile(path.Join(args.OutputDir, "favicon.ico"), favicon, 0666))
}

func HostRepo(data Repo) {
	if data.HostGit {
		debug("hosting of '%v' is ON", data.Name)
		old := path.Join(data.CloneDir(), ".git")
		new := path.Join(args.OutputDir, fmt.Sprintf("%v.git", data.Name))
		debug("renaming '%v', new %v", data.Name, new)
		checkErr(os.Rename(old, new))
		debug("running 'git update-server-info' in %v", new)
		cmd := exec.Command("git", "update-server-info")
		cmd.Dir = new
		checkErr(cmd.Run())
		debug("hosting '%v' at %v", data.Name, new)
	} else {
		debug("hosting of '%v' is OFF", data.Name)
	}
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

func checkFile(filename string) {
	_, err := os.Stat(filename)
	checkErr(err)
}

func checkDir(dir string) {
	_, err := os.ReadDir(dir)
	checkErr(err)
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

type CmdArgs struct {
	Silent     bool
	Wd         string
	ConfigPath string
	OutputDir  string
}

func DefaultCmdArgs() CmdArgs {
	return CmdArgs{
		Silent:     true,
		ConfigPath: "",
		OutputDir:  "",
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
