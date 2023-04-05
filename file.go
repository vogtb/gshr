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
	debug("file %v%v", f.RepoData.Name, f.Name)
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

func RenderSingleFilePages(data RepoData) {
	t, err := template.ParseFS(htmlTemplates, "templates/file.html", "templates/partials.html")
	checkErr(err)
	err = filepath.Walk(path.Join(args.CloneDir, data.Name), func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			_, canRenderExtension := settings.TextExtensions[ext]
			_, canRenderByFullName := settings.PlainFiles[filepath.Base(filename)]
			partialPath, _ := strings.CutPrefix(filename, path.Join(args.CloneDir, data.Name))
			outputName := path.Join(args.OutputDir, data.Name, "files", partialPath, "index.html")
			(&FilePage{
				RepoData:       data,
				Name:           partialPath,
				Extension:      ext,
				CanRender:      canRenderExtension || canRenderByFullName,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: path.Join(args.OutputDir, data.Name, "files", partialPath),
			}).Render(t)
		}
		return nil
	})
	checkErr(err)
}
