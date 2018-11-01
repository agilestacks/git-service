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
	Commit  string `json:"commit,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Date    string `json:"date,omitempty"`
	Author  string `json:"author,omitempty"`
	Subject string `json:"subject,omitempty"`
}

func Status(repoId, ref string) (*RepoStatus, error) {
	if ref == "" {
		ref = "master"
	}
	dir := filepath.Join(config.RepoDir, repoId)

	gitBin := gitBinPath()

	cmd := exec.Cmd{
		Path: gitBin,
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

	commit := commits[0]
	canonicalRef := refs[0]

	status := &RepoStatus{Commit: commit, Ref: canonicalRef}

	cmd = exec.Cmd{
		Path: gitBin,
		Dir:  dir,
		Args: []string{"git", "show", "-q", "--pretty=format:%aI%n%cn <%ce>%n%s", commit}, // iso date, author, message
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	outputBytes, err = cmd.Output()
	if len(outputBytes) == 0 {
		return status, fmt.Errorf("Git show `%s` returned no output", commit)
	}
	if err != nil {
		return status, err
	}
	output = string(outputBytes)
	lines = strings.Split(output, "\n")
	date := ""
	if len(lines) > 0 {
		date = lines[0]
	}
	author := ""
	if len(lines) > 1 {
		author = lines[1]
	}
	subject := ""
	if len(lines) > 2 {
		subject = lines[2]
	}
	return &RepoStatus{Commit: commit, Ref: canonicalRef, Author: author, Date: date, Subject: subject}, nil
}
