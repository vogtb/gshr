package main

import (
	"html/template"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitDetail struct {
	BaseURL         string
	Author          string
	AuthorEmail     string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
	LinesAdded      int
	LinesDeleted    int
}

func (c *CommitDetail) Render(t *template.Template) {
	err := os.MkdirAll(path.Join(config.OutputDir, "commit", c.Hash), 0755)
	checkErr(err)
	output, err := os.Create(path.Join(config.OutputDir, "commit", c.Hash, "index.html"))
	checkErr(err)
	err = t.Execute(output, c)
	checkErr(err)
}

func RenderAllCommitPages(r *git.Repository) {
	t, err := template.ParseFS(htmlTemplates, "templates/commit.html", "templates/partials.html")
	checkErr(err)
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
		commitDetail := CommitDetail{
			BaseURL:         config.BaseURL,
			Author:          c.Author.Name,
			AuthorEmail:     c.Author.Email,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
		}
		commitDetail.Render(t)
		return nil
	})
	checkErr(err)
}
