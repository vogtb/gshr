package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FileOverview struct {
	Mode   string
	Name   string
	Size   string
	Origin string
}

type FilesPage struct {
	RepoData repoData
	Files    []FileOverview
}

func (f *FilesPage) renderPage(t *template.Template) {
	debug("file page for '%v'", f.RepoData.Name)
	output, err := os.Create(path.Join(args.OutputDir, f.RepoData.Name, "files.html"))
	checkErr(err)
	err = t.Execute(output, f)
	checkErr(err)
}

func renderAllFilesPage(data repoData) {
	t, err := template.ParseFS(htmlTemplates, "template.files.html", "template.partials.html")
	checkErr(err)
	files := make([]FileOverview, 0)
	err = filepath.Walk(data.cloneDir(), func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			info, err := os.Stat(filename)
			checkErr(err)
			Name, _ := strings.CutPrefix(filename, data.cloneDir())
			Name, _ = strings.CutPrefix(Name, "/")
			tf := FileOverview{
				Origin: filename,
				Name:   Name,
				Mode:   info.Mode().String(),
				Size:   fmt.Sprintf("%v", info.Size()),
			}
			files = append(files, tf)
		}
		return nil
	})
	checkErr(err)
	index := FilesPage{
		RepoData: data,
		Files:    files,
	}
	index.renderPage(t)
}
