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

type FilePage struct {
	RepoData       repoData
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

func (f *FilePage) RenderPage(t *template.Template) {
	debug("file %v %v", f.RepoData.Name, f.Name)
	err := os.MkdirAll(f.DestinationDir, 0777)
	checkErr(err)
	err = os.MkdirAll(filepath.Dir(f.Destination), 0777)
	checkErr(err)
	output, err := os.Create(f.Destination)
	checkErr(err)
	err = t.Execute(output, f)
	checkErr(err)
}

func RenderSingleFilePages(data repoData) {
	t, err := template.ParseFS(htmlTemplates, "template.file.html", "template.partials.html")
	checkErr(err)
	err = filepath.Walk(data.cloneDir(), func(filename string, info fs.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			ext := filepath.Ext(filename)
			_, canRenderExtension := stt.TextExtensions[ext]
			_, canRenderByFullName := stt.PlainFiles[filepath.Base(filename)]
			canRender := canRenderExtension || canRenderByFullName
			partialPath, _ := strings.CutPrefix(filename, data.cloneDir())
			destDir := path.Join(args.OutputDir, data.Name, "files", partialPath)
			outputName := path.Join(args.OutputDir, data.Name, "files", partialPath, "index.html")
			var content template.HTML
			info, err := os.Stat(filename)
			checkErr(err)
			if canRender {
				fileBytes, err := os.ReadFile(filename)
				checkErr(err)
				fileStr := string(fileBytes)
				highlighted := highlight(destDir, &fileStr)
				checkErr(err)
				content = template.HTML(highlighted)
			}
			(&FilePage{
				RepoData:       data,
				Mode:           info.Mode().String(),
				Size:           fmt.Sprintf("%v", info.Size()),
				Name:           strings.TrimPrefix(partialPath, "/"),
				Extension:      ext,
				CanRender:      canRender,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: destDir,
				Content:        content,
			}).RenderPage(t)
		}
		return nil
	})
	checkErr(err)
}
