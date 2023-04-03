package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
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

func main() {
	outputDir = os.Getenv("OUTPUT_DIR")
	cloningDir = os.Getenv("CLONING_DIR")
	line("OUTPUT_DIR = %v", outputDir)
	line("CLONING_DIR = %v", cloningDir)
	cloneAndCheck()
	primaryScan()
}

func writeFile(filename string, outputFile string, t *template.Template) {
	fileBytes, err := os.ReadFile(filename)
	checkErr(err)
	fileStr := string(fileBytes)
	filepath.Dir(outputFile)
	err = os.MkdirAll(filepath.Dir(outputFile), 0775)
	checkErr(err)
	output, err := os.Create(outputFile)
	checkErr(err)
	err = t.Execute(output, fileStr)
	checkErr(err)
}

func cloneAndCheck() {
	r, err := git.PlainClone(cloningDir, false, &git.CloneOptions{
		URL: "/Users/bvogt/dev/src/ben/www",
	})
	checkErr(err)

	ref, err := r.Head()
	checkErr(err)

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)

	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	checkErr(err)

}

func primaryScan() {
	t, err := template.ParseFiles("template.html")
	checkErr(err)

	err = filepath.Walk(cloningDir, func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			line("READING: %v", filename)
			if _, ok := allowedExtensions[ext]; ok {
				partialPath, _ := strings.CutPrefix(filename, cloningDir)
				outputName := fmt.Sprintf("%v%v.html", outputDir, partialPath)
				line("WRITING: %v", outputName)
				writeFile(filename, outputName, t)
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
