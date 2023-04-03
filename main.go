package main

import (
	"bytes"
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
	debugOn        = true
	outputDir      = ""
	cloningDir     = ""
	textExtensions = map[string]bool{
		".conf":       true,
		".config":     true,
		".css":        true,
		".gitignore":  true,
		".gitmodules": true,
		".go":         true,
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
		".sum":        true,
		".toml":       true,
		".ts":         true,
		".tsx":        true,
		".txt":        true,
		".xml":        true,
		"Makefile":    true,
	}
)

func main() {
	flag.BoolVar(&debugOn, "debug", true, "debug mode")
	flag.StringVar(&outputDir, "output", "", "clone directory")
	flag.StringVar(&cloningDir, "clone", "", "clone directory")
	flag.Parse()

	if cloningDir == "" {
		cloningDir = fmt.Sprintf("/tmp/gshr-temp-clone-%v", rand.Uint32())
	}

	debug("output = %v", outputDir)
	debug("clone = %v", cloningDir)
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
	Destination    string
	DestinationDir string
	Content        template.HTML
}

func (f *TrackedFile) SaveTemplate(t *template.Template) {
	lexer := lexers.Match(f.DestinationDir)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("fruity")
	if style == nil {
		style = styles.Fallback
	}
	err := os.MkdirAll(f.DestinationDir, 0775)
	checkErr(err)
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
	err = os.MkdirAll(filepath.Dir(f.Destination), 0775)
	checkErr(err)
	output, err := os.Create(f.Destination)
	checkErr(err)
	f.Content = template.HTML(buf.String())
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
	output, err := os.Create(path.Join(outputDir, "log.html"))
	checkErr(err)
	err = t.Execute(output, mi)
	checkErr(err)
}

type FilesIndex struct {
	Files []TrackedFileMetaData
}

func (fi *FilesIndex) SaveTemplate(t *template.Template) {
	output, err := os.Create(path.Join(outputDir, "files.html"))
	checkErr(err)
	err = t.Execute(output, fi)
	checkErr(err)
}

func CloneAndInfo() *git.Repository {
	r, err := git.PlainClone(cloningDir, false, &git.CloneOptions{
		URL: "/Users/bvogt/dev/src/ben/www",
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
			Author:  c.Author.Email,
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
	err = filepath.Walk(cloningDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			if _, ok := textExtensions[ext]; ok {
				info, err := os.Stat(filename)
				checkErr(err)
				Name, _ := strings.CutPrefix(filename, cloningDir)
				Name, _ = strings.CutPrefix(Name, "/")
				tf := TrackedFileMetaData{
					Origin: filename,
					Name:   Name,
					Mode:   info.Mode().String(),
					Size:   fmt.Sprintf("%v", info.Size()),
				}
				trackedFiles = append(trackedFiles, tf)
			}
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

	err = filepath.Walk(cloningDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			if _, ok := textExtensions[ext]; ok {
				partialPath, _ := strings.CutPrefix(filename, cloningDir)
				outputName := path.Join(outputDir, "files", partialPath, "index.html")
				debug("reading = %v", partialPath)
				tf := TrackedFile{
					Origin:         filename,
					Destination:    outputName,
					DestinationDir: path.Join(outputDir, "files", partialPath),
				}
				tf.SaveTemplate(t)
			}
		}
		return nil
	})
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func debug(format string, a ...any) {
	if debugOn {
		fmt.Printf(format, a...)
		fmt.Print("\n")
	}
}
