package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gits/config"
)

func parseFlags() {
	var blobsFrom string
	var apiSecretEnvVar, hubApiSecretEnvVar, authApiSecretEnvVar string
	var hubApiEndpoint, hubApiHostEnvVar, hubApiPortEnvVar string
	var authApiEndpoint, authApiHostEnvVar, authApiPortEnvVar string

	flag.BoolVar(&config.Verbose, "verbose", true, "Print progress if set")
	flag.BoolVar(&config.Debug, "debug", false, "Print debug information if set")
	flag.BoolVar(&config.Trace, "trace", false, "Print detailed trace if set")
	flag.BoolVar(&config.ZipTrace, "zip_trace", false, "Print detailed debug of S3 ZIP streaming if set")

	flag.StringVar(&config.RepoDir, "repo_dir", "/git", "Base directory for Git repositories")
	flag.StringVar(&blobsFrom, "blobs", "", "Allowed URL prefixes to fetch repo sources from, empty for no restrictions")
	flag.IntVar(&config.HttpPort, "http_port", 8005, "HTTP API port to listen")
	flag.IntVar(&config.SshPort, "ssh_port", 2022, "SSH server port to listen")
	flag.StringVar(&config.HostKeyFile, "host_key", "gits-key", "Path to SSH server host private key file")
	flag.StringVar(&apiSecretEnvVar, "api_secret_env", "GIT_API_SECRET", "Environment variable to get secret from to protect Git HTTP API")

	flag.StringVar(&hubApiSecretEnvVar, "hub_api_secret_env", "HUB_API_SECRET", "Environment variable to get secret for Automation Hub HTTP API")
	flag.StringVar(&authApiSecretEnvVar, "auth_api_secret_env", "AUTH_API_SECRET", "Environment variable to get secret for Auth Service HTTP API")

	flag.BoolVar(&config.NoExtApiCalls, "no_ext_api_calls", false, "Emulate external calls to Automation Hub and Auth Service with internal stubs")
	flag.StringVar(&hubApiHostEnvVar, "hub_api_host_env", "HUB_SERVICE_HOST", "Environment variable to get Automation Hub HTTP API hostname / IP")
	flag.StringVar(&hubApiPortEnvVar, "hub_api_port_env", "HUB_SERVICE_PORT", "Environment variable to get Automation Hub HTTP API port")
	flag.StringVar(&authApiHostEnvVar, "auth_api_host_env", "AUTH_SERVICE_HOST", "Environment variable to get Auth Service HTTP API hostname / IP")
	flag.StringVar(&authApiPortEnvVar, "auth_api_port_env", "AUTH_SERVICE_PORT", "Environment variable to get Auth Service HTTP API port")

	flag.StringVar(&config.HubApiEndpoint, "hub_api", "", "Automation Hub HTTP API endpoint (overrides -hub_api_*_env / HUB_SERVICE_*)")
	flag.StringVar(&config.AuthApiEndpoint, "auth_api", "", "Auth Service HTTP API endpoint (overrides -auth_api_*_env / AUTH_SERVICE_*)")

	flag.StringVar(&config.AwsRegion, "aws_region", "", "The source archive bucket AWS region")
	flag.StringVar(&config.AwsProfile, "aws_profile", "", "The AWS credentials profile in ~/.aws/credentials")
	flag.BoolVar(&config.AwsUseIamRoleCredentials, "aws_use_iam_role_credentials", true, "Also search for EC2 instance credentials")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`Usage:
  gits -verbose -repo_dir /git -http_port 80 -ssh_port 22 -blobsFrom s3://agilestacks/

  If no -blobsFrom is set then any supported URL is allowed, currently s3://
  If -api_secret_env is empty or env variable exist but is empty then HTTP API is open - no access control.

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	config.GitApiSecret = lookupEnv(apiSecretEnvVar, "api_secret_env")
	if !config.NoExtApiCalls {
		config.HubApiSecret = lookupEnv(hubApiSecretEnvVar, "hub_api_secret_env")
		config.AuthApiSecret = lookupEnv(authApiSecretEnvVar, "auth_api_secret_env")

		config.HubApiEndpoint = lookupEndpoint(hubApiEndpoint, hubApiHostEnvVar, hubApiPortEnvVar, "hub")
		config.AuthApiEndpoint = lookupEndpoint(authApiEndpoint, authApiHostEnvVar, authApiPortEnvVar, "auth")
	}

	if blobsFrom != "" {
		config.BlobsFrom = strings.Split(blobsFrom, ",")
	}

	config.Update()
}

func lookupEnv(envVar string, param string) string {
	if envVar == "" {
		return ""
	}
	value, exist := os.LookupEnv(envVar)
	if !exist {
		log.Fatalf("`-%s %s` is set but variable not found in process environment, try `gits -h` or `gits -no_ext_api_calls`",
			param, envVar)
	}
	return value
}

func lookupEndpoint(endpoint string, hostEnv string, portEnv string, param string) string {
	if endpoint != "" {
		return endpoint
	}

	if hostEnv == "" {
		log.Fatal("-hub_api_host_env or -hub_api must be set")
	}
	host, exist := os.LookupEnv(hostEnv)
	if !exist || host == "" {
		log.Fatalf("-%s_api_host_env %s env var must point to API host", param, hostEnv)
	}

	port := "80"
	if portEnv != "" {
		port, exist := os.LookupEnv(portEnv)
		if !exist || port == "" {
			log.Fatalf("-%s_api_port_env %s env var must point to API port", param, hostEnv)
		}
	}

	proto := "http"
	if port == "443" {
		proto = "https"
	}
	if port == "80" || port == "443" {
		port = ""
	} else {
		port = ":" + port
	}

	return fmt.Sprintf("%s://%s%s", proto, host, port)
}
