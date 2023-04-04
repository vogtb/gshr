package main

import (
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type FilePage struct {
	RepoData       RepoData
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

func (f *FilePage) Render(t *template.Template) {
	err := os.MkdirAll(f.DestinationDir, 0775)
	checkErr(err)
	if f.CanRender {
		fileBytes, err := os.ReadFile(f.Origin)
		checkErr(err)
		fileStr := string(fileBytes)
		highlighted := highlight(f.DestinationDir, &fileStr)
		checkErr(err)
		f.Content = template.HTML(highlighted)
	}
	err = os.MkdirAll(filepath.Dir(f.Destination), 0775)
	checkErr(err)
	output, err := os.Create(f.Destination)
	checkErr(err)
	err = t.Execute(output, f)
	checkErr(err)
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
			(&FilePage{
				RepoData:       config.RepoData,
				Extension:      ext,
				CanRender:      canRenderExtension || canRenderByFullName,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: path.Join(config.OutputDir, "files", partialPath),
			}).Render(t)
		}
		return nil
	})
	checkErr(err)
}
