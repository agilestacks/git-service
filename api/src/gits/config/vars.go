package config

import (
	"log"
	"path/filepath"
	"strings"
)

const (
	GitBinDefault      = "/usr/bin/git"
	MultipartMaxMemory = 1024 * 1024
)

var (
	Verbose  bool
	Debug    bool
	Trace    bool

	RepoDir     string
	HttpPort    int
	SshPort     int
	HostKeyFile string
	BlobsFrom   []string

	GitApiSecret string

	NoExtApiCalls   bool
	HubApiSecret    string
	AuthApiSecret   string
	HubApiEndpoint  string
	AuthApiEndpoint string

	AwsRegion                string
	AwsProfile               string
	AwsUseIamRoleCredentials bool
)

func Update() {
	if Trace {
		Debug = true
	}
	if Debug {
		Verbose = true
	}
	for _, from := range BlobsFrom {
		if !strings.HasPrefix(from, "s3://") {
			log.Fatalf("Blob prefix `%s` not supported", from)
		}
	}
	if RepoDir == "" {
		RepoDir = "."
	} else if strings.HasPrefix(RepoDir, string(filepath.Separator)) {
		if len(RepoDir) == 1 {
			log.Fatalf("Repo directory `%s` is root directory?", RepoDir)
		}
	}
}
