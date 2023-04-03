package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

//go:embed file.template.html
var fileTemplateHtml string

//go:embed files.template.html
var filesTemplateHtml string

//go:embed log.template.html
var logTemplateHtml string

var (
	config Config
)

type Config struct {
	DebugOn        bool
	Repo           string
	OutputDir      string
	CloneDir       string
	BaseURL        string
	TextExtensions map[string]bool
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
			".css":        true,
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
			".toml":       true,
			".ts":         true,
			".tsx":        true,
			".txt":        true,
			".xml":        true,
			".yaml":       true,
			".yml":        true,
			"Makefile":    true,
		},
	}
}

func main() {
	config = DefaultConfig()
	flag.StringVar(&config.Repo, "repo", "", "Repo to use.")
	flag.BoolVar(&config.DebugOn, "debug", true, "Run in debug mode.")
	flag.StringVar(&config.OutputDir, "output", "", "Directory of output.")
	flag.StringVar(&config.CloneDir, "clone", "", "Directory to clone into. Random directory in /tmp if omitted.")
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
	RenderAllFilesPage()
	RenderSingleFilePages()
}

type TrackedFile struct {
	BaseURL        string
	Mode           string
	Name           string
	Size           string
	Origin         string
	Extension      string
	CanRender      bool
	Destination    string
	DestinationDir string
	Content        template.HTML
}

func (f *TrackedFile) Render(t *template.Template) {
	lexer := lexers.Match(f.DestinationDir)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("borland")
	if style == nil {
		style = styles.Fallback
	}
	err := os.MkdirAll(f.DestinationDir, 0775)
	checkErr(err)
	_, canRender := config.TextExtensions[f.Extension]
	if canRender {
		fileBytes, err := os.ReadFile(f.Origin)
		checkErr(err)
		fileStr := string(fileBytes)
		iterator, err := lexer.Tokenise(nil, fileStr)
		formatter := html.New(
			html.WithClasses(true),
			html.WithLineNumbers(true),
			html.LinkableLineNumbers(true, ""),
		)
		s := ""
		buf := bytes.NewBufferString(s)
		err = formatter.Format(buf, style, iterator)
		checkErr(err)
		f.Content = template.HTML(buf.String())
	}
	err = os.MkdirAll(filepath.Dir(f.Destination), 0775)
	checkErr(err)
	output, err := os.Create(f.Destination)
	checkErr(err)
	err = t.Execute(output, f)
	checkErr(err)
}

type Commit struct {
	Author          string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
}

type LogPage struct {
	BaseURL string
	Commits []Commit
}

func (mi *LogPage) Render(t *template.Template) {
	output, err := os.Create(path.Join(config.OutputDir, "log.html"))
	checkErr(err)
	err = t.Execute(output, mi)
	checkErr(err)
}

type TrackedFileMetaData struct {
	Mode   string
	Name   string
	Size   string
	Origin string
}

type FilesIndex struct {
	BaseURL string
	Files   []TrackedFileMetaData
}

func (fi *FilesIndex) Render(t *template.Template) {
	output, err := os.Create(path.Join(config.OutputDir, "files.html"))
	checkErr(err)
	err = t.Execute(output, fi)
	checkErr(err)
}

func CloneAndInfo() *git.Repository {
	r, err := git.PlainClone(config.CloneDir, false, &git.CloneOptions{
		URL: config.Repo,
	})
	checkErr(err)
	return r
}

func RenderLogPage(r *git.Repository) {
	t, err := template.New("log").Parse(logTemplateHtml)
	checkErr(err)
	commits := make([]Commit, 0)
	ref, err := r.Head()
	checkErr(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)

	err = cIter.ForEach(func(c *object.Commit) error {
		stats, err := c.Stats()
		checkErr(err)
		commits = append(commits, Commit{
			Author:          c.Author.Name,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
		})
		return nil
	})

	checkErr(err)
	m := LogPage{
		BaseURL: config.BaseURL,
		Commits: commits,
	}
	m.Render(t)
}

func RenderAllFilesPage() {
	t, err := template.New("files").Parse(filesTemplateHtml)
	checkErr(err)
	trackedFiles := make([]TrackedFileMetaData, 0)
	err = filepath.Walk(config.CloneDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			info, err := os.Stat(filename)
			checkErr(err)
			Name, _ := strings.CutPrefix(filename, config.CloneDir)
			Name, _ = strings.CutPrefix(Name, "/")
			tf := TrackedFileMetaData{
				Origin: filename,
				Name:   Name,
				Mode:   info.Mode().String(),
				Size:   fmt.Sprintf("%v", info.Size()),
			}
			trackedFiles = append(trackedFiles, tf)
		}
		return nil
	})
	checkErr(err)
	index := FilesIndex{
		BaseURL: config.BaseURL,
		Files:   trackedFiles,
	}
	index.Render(t)
}

func RenderSingleFilePages() {
	t, err := template.New("file").Parse(fileTemplateHtml)
	checkErr(err)
	err = filepath.Walk(config.CloneDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			_, canRender := config.TextExtensions[ext]
			partialPath, _ := strings.CutPrefix(filename, config.CloneDir)
			outputName := path.Join(config.OutputDir, "files", partialPath, "index.html")
			debug("reading = %v", partialPath)
			tf := TrackedFile{
				BaseURL:        config.BaseURL,
				Extension:      ext,
				CanRender:      canRender,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: path.Join(config.OutputDir, "files", partialPath),
			}
			tf.Render(t)
		}
		return nil
	})
	checkErr(err)
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
