package repo

import (
	"bytes"
	"encoding/hex"
	"fmt"
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

	isHash := guessIsCommitHash(ref)
	commit := ref
	canonicalRef := ""
	if !isHash {
		var err error
		commit, canonicalRef, err = commitByRef(dir, ref)
		if err != nil {
			return nil, err
		}
	}

	status := &RepoStatus{Commit: commit, Ref: canonicalRef}

	var stdoutBuffer bytes.Buffer
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "show", "-q", "--pretty=format:%aI%n%cn <%ce>%n%s", commit}, // iso date, author, message
	}
	gitDebug2(&cmd, &stdoutBuffer)
	err := cmd.Run()
	if stdoutBuffer.Len() == 0 {
		return status, fmt.Errorf("Git show `%s` returned no output", commit)
	}
	if err != nil {
		return status, err
	}
	output := stdoutBuffer.String()
	lines := strings.Split(output, "\n")
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

func guessIsCommitHash(ref string) bool {
	expected := hex.DecodedLen(len(ref))
	if expected != 20 {
		return false
	}
	dst := make([]byte, expected)
	decoded, err := hex.Decode(dst, []byte(ref))
	if err == nil && decoded == 20 {
		return true
	}
	return false
}

func commitByRef(dir, ref string) (string, string, error) {
	var stdoutBuffer bytes.Buffer
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "show-ref", ref},
	}
	gitDebug2(&cmd, &stdoutBuffer)
	err := cmd.Run()
	if stdoutBuffer.Len() == 0 {
		return "", "", fmt.Errorf("Ref `%s` not found", ref)
	}
	if err != nil {
		return "", "", err
	}
	output := stdoutBuffer.String()
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
		return "", "", fmt.Errorf("Ref `%s` not found", ref)
	}
	if len(refs) > 1 {
		return "", "", fmt.Errorf("Ref `%s` refer to multiple refs %v at %v", ref, refs, commits)
	}

	return commits[0], refs[0], nil
}
