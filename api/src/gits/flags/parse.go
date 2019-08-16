package flags

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gits/config"
)

func Parse() {
	var blobsFrom string
	var apiSecretEnvVar, hubApiSecretEnvVar, authApiSecretEnvVar, subsApiSecretEnvVar string
	var hubApiEndpoint, hubApiEndpointEnvVar, hubApiHostEnvVar, hubApiPortEnvVar string
	var authApiEndpoint, authApiEndpointEnvVar, authApiHostEnvVar, authApiPortEnvVar string
	var subsApiEndpoint, subsApiEndpointEnvVar, subsApiHostEnvVar, subsApiPortEnvVar string

	flag.BoolVar(&config.Verbose, "verbose", true, "Print progress if set")
	flag.BoolVar(&config.Debug, "debug", false, "Print debug information if set")
	flag.BoolVar(&config.Trace, "trace", false, "Print detailed trace if set")

	flag.StringVar(&config.RepoDir, "repo_dir", "/git", "Base directory for Git repositories")
	flag.StringVar(&config.MaintenanceFile, "maintenance", "", "Maintenance file, Git server go read-only mode if file exists (<repo_dir>/_maintenance)")
	flag.StringVar(&blobsFrom, "blobs", "", "Allowed URL prefixes to fetch repo sources from, empty for no restrictions")
	flag.IntVar(&config.HttpPort, "http_port", 8005, "HTTP API port to listen")
	flag.IntVar(&config.SshPort, "ssh_port", 2022, "SSH server port to listen")
	flag.StringVar(&config.HostKeyFile, "host_key", "gits-key", "Path to SSH server host private key file")
	flag.StringVar(&apiSecretEnvVar, "api_secret_env", "GIT_API_SECRET", "Environment variable to get secret from to protect Git HTTP API")

	flag.StringVar(&hubApiSecretEnvVar, "hub_api_secret_env", "HUB_API_SECRET", "Environment variable to get secret for Automation Hub HTTP API")
	flag.StringVar(&authApiSecretEnvVar, "auth_api_secret_env", "AUTH_API_SECRET", "Environment variable to get secret for Auth Service HTTP API")
	flag.StringVar(&subsApiSecretEnvVar, "subs_api_secret_env", "SUBS_API_SECRET", "Environment variable to get secret for Subscriptions Service HTTP API")

	flag.BoolVar(&config.NoExtApiCalls, "no_ext_api_calls", false, "Emulate external calls to Automation Hub and Auth Service with internal stubs")
	flag.StringVar(&hubApiEndpointEnvVar, "hub_api_endpoint_env", "HUB_SERVICE_ENDPOINT", "Environment variable to get Automation Hub HTTP API endpoint")
	flag.StringVar(&hubApiHostEnvVar, "hub_api_host_env", "HUB_SERVICE_HOST", "Environment variable to get Automation Hub HTTP API hostname / IP")
	flag.StringVar(&hubApiPortEnvVar, "hub_api_port_env", "HUB_SERVICE_PORT", "Environment variable to get Automation Hub HTTP API port")
	flag.StringVar(&authApiEndpointEnvVar, "auth_api_endpoint_env", "AUTH_SERVICE_ENDPOINT", "Environment variable to get Auth Service HTTP API endpoint")
	flag.StringVar(&authApiHostEnvVar, "auth_api_host_env", "AUTH_SERVICE_HOST", "Environment variable to get Auth Service HTTP API hostname / IP")
	flag.StringVar(&authApiPortEnvVar, "auth_api_port_env", "AUTH_SERVICE_PORT", "Environment variable to get Auth Service HTTP API port")
	flag.StringVar(&subsApiEndpointEnvVar, "subs_api_endpoint_env", "SUBS_SERVICE_ENDPOINT", "Environment variable to get Subscriptions Service HTTP API endpoint")
	flag.StringVar(&subsApiHostEnvVar, "subs_api_host_env", "SUBS_SERVICE_HOST", "Environment variable to get Subscriptions Service HTTP API hostname / IP")
	flag.StringVar(&subsApiPortEnvVar, "subs_api_port_env", "SUBS_SERVICE_PORT", "Environment variable to get Subscriptions Service HTTP API port")

	flag.StringVar(&hubApiEndpoint, "hub_api_endpoint", "", "Automation Hub HTTP API endpoint (overrides -hub_api_*_env / HUB_SERVICE_*)")
	flag.StringVar(&authApiEndpoint, "auth_api_endpoint", "", "Auth Service HTTP API endpoint (overrides -auth_api_*_env / AUTH_SERVICE_*)")
	flag.StringVar(&subsApiEndpoint, "subs_api_endpoint", "", "Subscriptions Service HTTP API endpoint (overrides -subs_api_*_env / SUBS_SERVICE_*)")

	flag.StringVar(&config.AwsRegion, "aws_region", "", "The source archive bucket AWS region")
	flag.StringVar(&config.AwsProfile, "aws_profile", "", "The AWS credentials profile of ~/.aws/credentials")
	flag.BoolVar(&config.AwsUseIamRoleCredentials, "aws_use_iam_role_credentials", true, "Try EC2 instance credentials")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`Usage:
  gits -verbose -repo_dir /git -http_port 80 -ssh_port 22 -blobs s3://agilestacks-distribution/ -aws_region us-east-2

  If no -blobs is set then any supported URL is allowed, currently s3://
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
		config.SubsApiSecret = lookupEnv(subsApiSecretEnvVar, "subs_api_secret_env")

		config.HubApiEndpoint = lookupEndpoint(hubApiEndpoint, hubApiEndpointEnvVar, hubApiHostEnvVar, hubApiPortEnvVar, "hub")
		config.AuthApiEndpoint = lookupEndpoint(authApiEndpoint, authApiEndpointEnvVar, authApiHostEnvVar, authApiPortEnvVar, "auth")
		config.SubsApiEndpoint = lookupEndpoint(subsApiEndpoint, subsApiEndpointEnvVar, subsApiHostEnvVar, subsApiPortEnvVar, "subs")
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

func lookupEndpoint(endpoint string, endpointEnv string, hostEnv string, portEnv string, param string) string {
	if endpoint != "" {
		return endpoint
	}

	if endpointEnv != "" {
		endpoint, _ := os.LookupEnv(endpointEnv)
		if endpoint != "" {
			return endpoint
		}
	}

	if hostEnv != "" {
		host, _ := os.LookupEnv(hostEnv)
		if host != "" {
			port := "80"
			if portEnv != "" {
				var exist bool
				port, exist = os.LookupEnv(portEnv)
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
	}

	log.Fatalf("One of -%[1]s_api_endpoint, -%[1]s_api_endpoint_env, -%[1]s_api_host_env must be set", param)
	return ""
}
