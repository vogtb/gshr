package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var allowedExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".html": true,
	".css":  true,
	".js":   true,
	".toml": true,
	".json": true,
	".lock": true,
}

var (
	outputDir  = ""
	cloningDir = ""
)

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
}

func (tf *TrackedFile) SaveTemplate(t *template.Template) {
	err := os.MkdirAll(tf.DestinationDir, 0775)
	checkErr(err)
	fileBytes, err := os.ReadFile(tf.Origin)
	checkErr(err)
	fileStr := string(fileBytes)
	err = os.MkdirAll(filepath.Dir(tf.Destination), 0775)
	checkErr(err)
	output, err := os.Create(tf.Destination)
	checkErr(err)
	err = t.Execute(output, fileStr)
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

func main() {
	outputDir = os.Getenv("OUTPUT_DIR")
	cloningDir = os.Getenv("CLONING_DIR")
	line("OUTPUT_DIR = %v", outputDir)
	line("CLONING_DIR = %v", cloningDir)
	r := CloneAndInfo()
	BuildLogPage(r)
	BuildFilesPages()
	BuildSingleFilePages()
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
		fmt.Println(c)
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
			if _, ok := allowedExtensions[ext]; ok {
				info, err := os.Stat(filename)
				checkErr(err)
				Name, _ := strings.CutPrefix(filename, cloningDir)
				Name, _ = strings.CutPrefix(Name, "/")
				tf := TrackedFileMetaData{
					Origin: filename,
					Name:   Name,
					Mode:   info.Mode().String(),
					Size:   fmt.Sprintf("%v bytes", info.Size()),
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
			if _, ok := allowedExtensions[ext]; ok {
				partialPath, _ := strings.CutPrefix(filename, cloningDir)
				outputName := path.Join(outputDir, "files", partialPath, "index.html")
				line("READING: %v", filename)
				line("WRITING: %v", outputName)
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

func line(format string, a ...any) {
	fmt.Printf(format, a...)
	fmt.Print("\n")
}
