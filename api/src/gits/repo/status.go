package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
)

type RepoStatus struct {
	Commit string `json:"commit"`
	Ref    string `json:"ref"`
}

func Status(repoId, ref string) (*RepoStatus, error) {
	if ref == "" {
		ref = "master"
	}
	dir := filepath.Join(config.RepoDir, repoId)
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "show-ref", ref},
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	outputBytes, err := cmd.Output()
	if len(outputBytes) == 0 {
		return nil, fmt.Errorf("Ref `%s` not found", ref)
	}
	if err != nil {
		return nil, err
	}
	output := string(outputBytes)
	lines := strings.Split(output, "\n")
	commits := make([]string, 0, len(lines))
	refs := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		commits = append(commits, parts[0])
		refs = append(refs, parts[1])
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("Ref `%s` not found", ref)
	}
	if len(refs) > 1 {
		return nil, fmt.Errorf("Ref `%s` refer to multiple refs %v at %v", ref, refs, commits)
	}
	return &RepoStatus{Commit: commits[0], Ref: refs[0]}, nil
}
