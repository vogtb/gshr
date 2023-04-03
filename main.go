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

type TrackedFile struct {
	origin      string
	destination string
}

func (tf *TrackedFile) SaveTemplate(t *template.Template) {
	fileBytes, err := os.ReadFile(tf.origin)
	checkErr(err)
	fileStr := string(fileBytes)
	err = os.MkdirAll(filepath.Dir(tf.destination), 0775)
	checkErr(err)
	output, err := os.Create(tf.destination)
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

type MainIndex struct {
	Commits []GshrCommit
}

func (mi *MainIndex) SaveTemplate(t *template.Template) {
	output, err := os.Create(path.Join(outputDir, "index.html"))
	checkErr(err)
	err = t.Execute(output, mi)
	checkErr(err)
}

func main() {
	outputDir = os.Getenv("OUTPUT_DIR")
	cloningDir = os.Getenv("CLONING_DIR")
	line("OUTPUT_DIR = %v", outputDir)
	line("CLONING_DIR = %v", cloningDir)
	r := CloneAndInfo()
	BuildMainIndex(r)
	BuildTrackedFiles()
}

func CloneAndInfo() *git.Repository {
	r, err := git.PlainClone(cloningDir, false, &git.CloneOptions{
		URL: "/Users/bvogt/dev/src/ben/www",
	})
	checkErr(err)
	return r
}

func BuildMainIndex(r *git.Repository) {
	t, err := template.ParseFiles("index.template.html")
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
	m := MainIndex{
		Commits: commits,
	}
	m.SaveTemplate(t)

}

func BuildTrackedFiles() {
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
				outputName := fmt.Sprintf("%v%v.html", outputDir, partialPath)
				line("READING: %v", filename)
				line("WRITING: %v", outputName)
				tf := TrackedFile{
					origin:      filename,
					destination: outputName,
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
