package main

import (
	"html/template"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type LogPageCommit struct {
	Author          string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
	LinesAdded      int
	LinesDeleted    int
}

type LogPage struct {
	RepoData   repoData
	HasReadMe  bool
	ReadMePath string
	Commits    []LogPageCommit
}

func (l *LogPage) renderPage(t *template.Template) {
	debug("log page for '%v'", l.RepoData.Name)
	output, err := os.Create(path.Join(args.OutputDir, l.RepoData.Name, "log.html"))
	checkErr(err)
	err = t.Execute(output, l)
	checkErr(err)
}

func renderLogPage(data repoData, r *git.Repository) {
	t, err := template.ParseFS(htmlTemplates, "template.log.html", "template.partials.html")
	checkErr(err)
	commits := make([]LogPageCommit, 0)
	ref, err := r.Head()
	checkErr(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	checkErr(err)
	err = cIter.ForEach(func(c *object.Commit) error {
		stats, err := c.Stats()
		added := 0
		deleted := 0
		for i := 0; i < len(stats); i++ {
			stat := stats[i]
			added += stat.Addition
			deleted += stat.Deletion
		}
		checkErr(err)
		commits = append(commits, LogPageCommit{
			Author:          c.Author.Name,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
		})
		return nil
	})
	checkErr(err)
	(&LogPage{
		RepoData: data,
		Commits:  commits,
	}).renderPage(t)
}
