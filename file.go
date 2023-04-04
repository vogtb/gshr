package main

import (
	"bytes"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type TrackedFile struct {
	BaseURL        string
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

func (f *TrackedFile) Render(t *template.Template) {
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
	if f.CanRender {
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
			tf := TrackedFile{
				BaseURL:        config.BaseURL,
				Extension:      ext,
				CanRender:      canRenderExtension || canRenderByFullName,
				Origin:         filename,
				Destination:    outputName,
				DestinationDir: path.Join(config.OutputDir, "files", partialPath),
			}
			tf.Render(t)
		}
		return nil
	})
	checkErr(err)
}
