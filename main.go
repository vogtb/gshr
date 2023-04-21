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
	"time"

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

var args cmdArgs

var conf config

var stt settings

func main() {
	var r *git.Repository = &git.Repository{}
	initialize()
	allRepoData := []repoData{}
	for _, repo := range conf.Repos {
		data := cloneAndGetData(repo, r)
		allRepoData = append(allRepoData, data)
		renderLogPage(data, r)
		renderAllCommitPages(data, r)
		renderAllFilesPage(data)
		renderIndividualFilePages(data)
	}
	renderIndexPage(allRepoData)
	renderAssets()
	for _, repo := range conf.Repos {
		hostRepo(repo)
	}
}

func initialize() {
	log.SetFlags(0)
	log.SetOutput(new(logger))
	args = defaultCmdArgs()
	stt = defaultSettings()
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
	conf = parseConfig(configString)
	debug("base_url '%v'", conf.Site.BaseURL)
	debug("site_name '%v'", conf.Site.Name)
	conf.validate()
}

func cloneAndGetData(repo repoConfig, r *git.Repository) repoData {
	err := os.MkdirAll(repo.cloneDir(), 0755)
	checkErr(err)
	err = os.MkdirAll(path.Join(args.OutputDir, repo.Name), 0755)
	checkErr(err)
	debug("cloning '%v'", repo.Name)
	repoRef, err := git.PlainClone(repo.cloneDir(), false, &git.CloneOptions{
		URL: repo.URL,
	})
	checkErr(err)
	data := repoData{
		repoConfig: repo,
		AltLink:    repo.AltLink,
		BaseURL:    conf.Site.BaseURL,
		HeadData: HeadData{
			BaseURL:  conf.Site.BaseURL,
			SiteName: conf.Site.Name,
			GenTime:  args.GenTime,
		},
		ReadMePath:      repo.findFileInRoot(stt.AllowedReadMeFiles),
		LicenseFilePath: repo.findFileInRoot(stt.AllowedLicenseFiles),
	}
	*r = *repoRef
	return data
}

func renderAssets() {
	debug("rendering gshr.css")
	debug("rendering favicon.ico")
	checkErr(os.WriteFile(path.Join(args.OutputDir, "gshr.css"), css, 0666))
	checkErr(os.WriteFile(path.Join(args.OutputDir, "favicon.ico"), favicon, 0666))
}

func hostRepo(data repoConfig) {
	debug("hosting '%v'", data.Name)
	old := path.Join(data.cloneDir(), ".git")
	renamed := path.Join(args.OutputDir, fmt.Sprintf("%v.git", data.Name))
	repoFiles := path.Join(args.OutputDir, data.Name, "git")
	final := path.Join(args.OutputDir, fmt.Sprintf("%v.git", data.Name))
	debug("renaming '%v' to %v", data.Name, renamed)
	checkErr(os.Rename(old, renamed))
	debug("running 'git update-server-info' in %v", renamed)
	cmd := exec.Command("git", "update-server-info")
	cmd.Dir = renamed
	checkErr(cmd.Run())
	os.RemoveAll(repoFiles)
	debug("hosting '%v' at %v", data.Name, final)
}

type logger struct{}

func (writer logger) Write(bytes []byte) (int, error) {
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

type cmdArgs struct {
	Silent     bool
	Wd         string
	ConfigPath string
	OutputDir  string
	GenTime    string
}

func defaultCmdArgs() cmdArgs {
	return cmdArgs{
		Silent:     true,
		ConfigPath: "",
		OutputDir:  "",
		GenTime:    time.Now().Format("Mon Jan 2 15:04:05 MST 2006"),
	}
}

type settings struct {
	TextExtensions      map[string]bool
	PlainFiles          map[string]bool
	AllowedLicenseFiles map[string]bool
	AllowedReadMeFiles  map[string]bool
}

func defaultSettings() settings {
	return settings{
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
