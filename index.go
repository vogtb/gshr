package main

import (
	"html/template"
	"os"
	"path"
)

type IndexPage struct {
	BaseURL  string
	SiteName string
	Repos    []RepoData
}

func (l *IndexPage) Render(t *template.Template) {
	debug("index for '%v'", l.SiteName)
	output, err := os.Create(path.Join(args.OutputDir, "index.html"))
	checkErr(err)
	err = t.Execute(output, l)
	checkErr(err)
}

func RenderIndexPage(repos []RepoData) {
	t, err := template.ParseFS(htmlTemplates, "templates/index.html", "templates/partials.html")
	checkErr(err)
	(&IndexPage{
		BaseURL:  config.BaseURL,
		SiteName: config.SiteName,
		Repos:    repos,
	}).Render(t)
}
