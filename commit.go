package main

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type FileDiff struct {
}

type CommitPage struct {
	RepoData        RepoData
	Author          string
	AuthorEmail     string
	Date            string
	Hash            string
	Message         string
	FileChangeCount int
	LinesAdded      int
	LinesDeleted    int
	DiffContent     template.HTML
}

func (c *CommitPage) Render(t *template.Template) {
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
		parent, err := c.Parent(0)
		if err != nil && errors.Is(err, object.ErrParentNotFound) {
			// ok
		} else if err != nil {
			checkErr(err)
		}
		diffContent := template.HTML("")
		if parent != nil {
			lexer := lexers.Match("x.diff")
			if lexer == nil {
				lexer = lexers.Fallback
			}
			style := styles.Get("borland")
			if style == nil {
				style = styles.Fallback
			}
			patch, err := c.Patch(parent)
			checkErr(err)
			patchString := patch.String()
			iterator, err := lexer.Tokenise(nil, patchString)
			formatter := html.New(
				html.WithClasses(true),
				html.WithLineNumbers(true),
				html.LinkableLineNumbers(true, ""),
			)
			s := ""
			buf := bytes.NewBufferString(s)
			err = formatter.Format(buf, style, iterator)
			checkErr(err)
			diffContent = template.HTML(buf.String())
		}
		stats, err := c.Stats()
		added := 0
		deleted := 0
		for i := 0; i < len(stats); i++ {
			stat := stats[i]
			added += stat.Addition
			deleted += stat.Deletion
		}
		checkErr(err)
		(&CommitPage{
			RepoData:        config.RepoData,
			Author:          c.Author.Name,
			AuthorEmail:     c.Author.Email,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
			DiffContent:     diffContent,
		}).Render(t)
		return nil
	})
	checkErr(err)
}
