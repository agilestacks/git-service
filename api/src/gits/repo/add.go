package repo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v3/memfs"
	"gopkg.in/src-d/go-git.v4"
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

	options := &git.CloneOptions{
		URL:          dir,
		SingleBranch: true,
		Depth:        1,
		Progress:     progress,
	}
	clone, err := git.Clone(storer, fs, options)
	if err != nil {
		return fmt.Errorf("Unable to clone Git repo `%s`: %v", dir, err)
	}
	worktree, err := clone.Worktree()
	if err != nil {
		return fmt.Errorf("Unable to open Git repo in-memory worktree: %v", dir, err)
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
	_, err = worktree.Commit(commitMessage, &git.CommitOptions{})
	if err != nil {
		return err
	}

	pushOptions := git.PushOptions{RemoteName: "origin"}
	err = clone.Push(&pushOptions)
	if err != nil {
		return err
	}

	return nil
}
