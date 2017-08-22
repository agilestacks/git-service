package repo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/src-d/go-billy.v3/memfs"
	"gopkg.in/src-d/go-git.v4"
	plumbingObject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"gits/config"
)

type AddFile struct {
	Path    string
	Content io.Reader
}

func Add(repoId string, files []AddFile, commitMessage string) error {
	dir := filepath.Join(config.RepoDir, repoId)

	storer := memory.NewStorage()
	fs := memfs.New()

	progress := os.Stdout
	if !config.Debug {
		progress = nil
	}

	cloneOptions := &git.CloneOptions{
		URL:          dir,
		SingleBranch: true,
		Depth:        1,
		Progress:     progress,
	}
	clone, err := git.Clone(storer, fs, cloneOptions)
	if err != nil {
		return fmt.Errorf("Unable to clone Git repo `%s`: %v", dir, err)
	}
	worktree, err := clone.Worktree()
	if err != nil {
		return fmt.Errorf("Unable to open Git repo %s in-memory worktree: %v", dir, err)
	}

	for _, file := range files {
		fileDir := filepath.Dir(file.Path)
		if fileDir != "." {
			err := fs.MkdirAll(fileDir, dirMode)
			if err != nil {
				return err
			}
		}
		out, err := fs.Create(file.Path)
		if err != nil {
			return err
		}
		io.Copy(out, file.Content)
		out.Close()
		worktree.Add(file.Path)
	}

	commitOptions := git.CommitOptions{
		Author: &plumbingObject.Signature{
			Name:  "Automation Hub",
			Email: "hub@agilestack.com",
			When:  time.Now(),
		},
	}
	_, err = worktree.Commit(commitMessage, &commitOptions)
	if err != nil {
		return err
	}

	pushOptions := git.PushOptions{RemoteName: "origin"}
	err = clone.Push(&pushOptions)
	if err != nil {
		return err
	}

	if config.Verbose {
		added := make([]string, 0, len(files))
		for _, file := range files {
			added = append(added, file.Path)
		}
		log.Printf("Added `%s` to `%s`", strings.Join(added, ", "), repoId)
	}

	return nil
}
