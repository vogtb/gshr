package main

import (
	"bytes"
	"embed"
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
	if f.CanRender {
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
	BaseURL         string
	Author          string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
	LinesAdded      int
	LinesDeleted    int
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
	BaseURL string
	Mode    string
	Name    string
	Size    string
	Origin  string
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

type CommitDetail struct {
	BaseURL         string
	Author          string
	AuthorEmail     string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
	LinesAdded      int
	LinesDeleted    int
}

func (c *CommitDetail) Render(t *template.Template) {
	err := os.MkdirAll(path.Join(config.OutputDir, "commit", c.Hash), 0755)
	checkErr(err)
	output, err := os.Create(path.Join(config.OutputDir, "commit", c.Hash, "index.html"))
	checkErr(err)
	err = t.Execute(output, c)
	checkErr(err)
}

func CloneAndInfo() *git.Repository {
	r, err := git.PlainClone(config.CloneDir, false, &git.CloneOptions{
		URL: config.Repo,
	})
	checkErr(err)
	return r
}

func RenderAllCommitPages(r *git.Repository) {
	t, err := template.ParseFS(htmlTemplates, "templates/commit.html", "templates/partials.html")
	checkErr(err)
	ref, err := r.Head()
	checkErr(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)
	err = cIter.ForEach(func(c *object.Commit) error {
		stats, err := c.Stats()
		added := 0
		deleted := 0
		for i := 0; i < len(stats); i++ {
			stat := stats[i]
			added += stat.Addition
			deleted += stat.Deletion
		}
		checkErr(err)
		commitDetail := CommitDetail{
			BaseURL:         config.BaseURL,
			Author:          c.Author.Name,
			AuthorEmail:     c.Author.Email,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
		}
		commitDetail.Render(t)
		return nil
	})
	checkErr(err)
}

func RenderLogPage(r *git.Repository) {
	t, err := template.ParseFS(htmlTemplates, "templates/log.html", "templates/partials.html")
	checkErr(err)
	commits := make([]Commit, 0)
	ref, err := r.Head()
	checkErr(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)

	err = cIter.ForEach(func(c *object.Commit) error {
		stats, err := c.Stats()
		added := 0
		deleted := 0
		for i := 0; i < len(stats); i++ {
			stat := stats[i]
			added += stat.Addition
			deleted += stat.Deletion
		}
		checkErr(err)
		commits = append(commits, Commit{
			BaseURL:         config.BaseURL,
			Author:          c.Author.Name,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
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
	t, err := template.ParseFS(htmlTemplates, "templates/files.html", "templates/partials.html")
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
				BaseURL: config.BaseURL,
				Origin:  filename,
				Name:    Name,
				Mode:    info.Mode().String(),
				Size:    fmt.Sprintf("%v", info.Size()),
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
	t, err := template.ParseFS(htmlTemplates, "templates/file.html", "templates/partials.html")
	checkErr(err)
	err = filepath.Walk(config.CloneDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			_, canRenderExtension := config.TextExtensions[ext]
			_, canRenderByFullName := config.PlainFiles[filepath.Base(filename)]
			partialPath, _ := strings.CutPrefix(filename, config.CloneDir)
			outputName := path.Join(config.OutputDir, "files", partialPath, "index.html")
			debug("reading = %v", partialPath)
			tf := TrackedFile{
				BaseURL:        config.BaseURL,
				Extension:      ext,
				CanRender:      canRenderExtension || canRenderByFullName,
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
