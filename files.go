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
