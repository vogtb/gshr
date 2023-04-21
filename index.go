package main

import (
	"html/template"
	"os"
	"path"
)

type IndexPage struct {
	HeadData HeadData
	Repos    []repoData
}

func (l *IndexPage) renderPage(t *template.Template) {
	debug("index for '%v'", l.HeadData.SiteName)
	output, err := os.Create(path.Join(args.OutputDir, "index.html"))
	checkErr(err)
	err = t.Execute(output, l)
	checkErr(err)
}

func renderIndexPage(repos []repoData) {
	t, err := template.ParseFS(htmlTemplates, "template.index.html", "template.partials.html")
	checkErr(err)
	(&IndexPage{
		HeadData: HeadData{
			BaseURL:  conf.Site.BaseURL,
			SiteName: conf.Site.Name,
			GenTime:  args.GenTime,
		},
		Repos: repos,
	}).renderPage(t)
}
