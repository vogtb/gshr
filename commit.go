package main

import (
	"errors"
	"html/template"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

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
	FilesChanged    []string
	DiffContent     template.HTML
}

func (c *CommitPage) RenderPage(t *template.Template) {
	debug("commit %v %v", c.RepoData.Name, c.Hash)
	err := os.MkdirAll(path.Join(args.OutputDir, c.RepoData.Name, "commit", c.Hash), 0755)
	checkErr(err)
	output, err := os.Create(path.Join(args.OutputDir, c.RepoData.Name, "commit", c.Hash, "index.html"))
	checkErr(err)
	err = t.Execute(output, c)
	checkErr(err)
}

func RenderAllCommitPages(data RepoData, r *git.Repository) {
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
		filesChangedMap := make(map[string]bool)
		filesChanged := []string{}
		if parent != nil {

			patch, err := c.Patch(parent)
			checkErr(err)
			patchString := patch.String()
			highlighted := highlight("x.diff", &patchString)
			checkErr(err)
			diffContent = template.HTML(highlighted)
			for _, fp := range patch.FilePatches() {
				from, to := fp.Files()
				if from != nil {
					filePath := from.Path()
					if _, found := filesChangedMap[filePath]; !found {
						filesChangedMap[filePath] = true
						filesChanged = append(filesChanged, filePath)
					}
				}
				if to != nil {
					filePath := to.Path()
					if _, found := filesChangedMap[filePath]; !found {
						filesChangedMap[filePath] = true
						filesChanged = append(filesChanged, filePath)
					}
				}
			}
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
			RepoData:        data,
			Author:          c.Author.Name,
			AuthorEmail:     c.Author.Email,
			Message:         c.Message,
			Date:            c.Author.When.UTC().Format("2006-01-02 15:04:05"),
			Hash:            c.Hash.String(),
			FileChangeCount: len(stats),
			LinesAdded:      added,
			LinesDeleted:    deleted,
			FilesChanged:    filesChanged,
			DiffContent:     diffContent,
		}).RenderPage(t)
		return nil
	})
	checkErr(err)
}
