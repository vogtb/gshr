package main

import (
	"bytes"
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

var (
	config Config
)

type Config struct {
	DebugOn        bool
	Repo           string
	OutputDir      string
	CloneDir       string
	TextExtensions map[string]bool
}

func DefaultConfig() Config {
	return Config{
		DebugOn:   true,
		Repo:      "",
		OutputDir: "",
		CloneDir:  fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32()),
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
	flag.Parse()

	if config.Repo == "" {
		checkErr(errors.New("--repo flag is required"))
	}

	debug("output = %v", config.OutputDir)
	debug("clone = %v", config.CloneDir)
	r := CloneAndInfo()
	BuildLogPage(r)
	BuildFilesPages()
	BuildSingleFilePages()
}

type TrackedFileMetaData struct {
	Mode   string
	Name   string
	Size   string
	Origin string
}

type TrackedFile struct {
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

func (f *TrackedFile) SaveTemplate(t *template.Template) {
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

type GshrCommit struct {
	Author  string
	Date    string
	Hash    string
	Message string
}

type LogPage struct {
	Commits []GshrCommit
}

func (mi *LogPage) SaveTemplate(t *template.Template) {
	output, err := os.Create(path.Join(config.OutputDir, "log.html"))
	checkErr(err)
	err = t.Execute(output, mi)
	checkErr(err)
}

type FilesIndex struct {
	Files []TrackedFileMetaData
}

func (fi *FilesIndex) SaveTemplate(t *template.Template) {
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

func BuildLogPage(r *git.Repository) {
	t, err := template.ParseFiles("log.template.html")
	commits := make([]GshrCommit, 0)
	ref, err := r.Head()
	checkErr(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)

	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, GshrCommit{
			Author:  c.Author.Name,
			Message: c.Message,
			Date:    c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:    c.Hash.String(),
		})
		return nil
	})

	checkErr(err)
	m := LogPage{
		Commits: commits,
	}
	m.SaveTemplate(t)
}

func BuildFilesPages() {
	t, err := template.ParseFiles("files.template.html")
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
		Files: trackedFiles,
	}
	index.SaveTemplate(t)
}

func BuildSingleFilePages() {
	t, err := template.ParseFiles("file.template.html")
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
				Extension:      ext,
				CanRender:      canRender,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: path.Join(config.OutputDir, "files", partialPath),
			}
			tf.SaveTemplate(t)
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
