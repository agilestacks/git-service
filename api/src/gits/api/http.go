package api

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gits/config"
)

const (
	apiRepoHandlerPath = "/api/v1/repositories/"
	gitHandlerPath     = "/repo/"
)

func Listen(host string, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc(apiRepoHandlerPath, apiRepositoriesRouter)
	mux.HandleFunc(gitHandlerPath, gitRouter)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go listen(server)
}

func listen(server *http.Server) {
	log.Fatalf("Error in HTTP server: %v", server.ListenAndServe())
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json := fmt.Sprintf("{ \"error\": \"%s\" }", strings.Replace(message, "\"", "\\\"", -1))
	io.WriteString(w, json)
}

func apiRepositoriesRouter(w http.ResponseWriter, req *http.Request) {
	if !rightApiSecret(w, req) {
		return
	}

	repoId, verb, path, err := parsePath(req.URL.Path, apiRepoHandlerPath)
	if err != nil {
		message := fmt.Sprintf("Error parsing request path %q: %v", req.URL.Path, err)
		if config.Verbose {
			log.Print(message)
		}
		writeError(w, http.StatusBadRequest, message)
		return
	}

	handled := false
	switch req.Method {

	default:
		break

	case "GET":
		if verb == "log" && path == "" {
			sendRepoLog(repoId, w)
			handled = true
		}
		break

	case "PUT":
		if verb == "" {
			createRepo(repoId, req.Body, w)
			handled = true
		} else if verb == "commit" && path != "" {
			uploadFile(repoId, path, req, w)
			handled = true
		}
		break

	case "POST":
		if verb == "commit" && path == "" {
			uploadFiles(repoId, req, w)
			handled = true
		}
		break

	case "DELETE":
		if verb == "" {
			deleteRepo(repoId, w)
			handled = true
		}
	}

	if !handled {
		if config.Verbose {
			log.Printf("Cannot route HTTP %q request on repo %q; verb %q; path %q", req.Method, repoId, verb, path)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func gitRouter(w http.ResponseWriter, req *http.Request) {
	if !rightApiSecret(w, req) {
		return
	}

	repoId, verb, path, err := parsePath(req.URL.Path, gitHandlerPath)
	if err != nil {
		message := fmt.Sprintf("Error parsing request path %q: %v", req.URL.Path, err)
		if config.Verbose {
			log.Print(message)
		}
		writeError(w, http.StatusBadRequest, message)
		return
	}

	handled := false
	switch req.Method {

	default:
		break

	case "GET":
		service := req.URL.Query().Get("service")
		if verb == "info" && path == "refs" && service == "git-upload-pack" {
			sendRefsInfo(repoId, w)
			handled = true
		}
		break

	case "POST":
		if verb == "git-upload-pack" {
			in := req.Body
			if req.Header.Get("Content-Encoding") == "gzip" {
				in, err = gzip.NewReader(in)
				if err != nil {
					log.Printf("Unable to decompress request: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			sendRefsPack(repoId, w, in)
			in.Close()
			handled = true
		}
		break
	}

	if !handled {
		if config.Verbose {
			log.Printf("Cannot route HTTP %q request on repo %q; verb %q; path %q", req.Method, repoId, verb, path)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func rightApiSecret(w http.ResponseWriter, req *http.Request) bool {
	if config.GitApiSecret != "" {
		secret := req.Header.Get("X-API-Secret")
		if secret != config.GitApiSecret {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
	}
	return true
}

func parsePath(urlPath string, prefixMust string) (string, string, string, error) {
	if !strings.HasPrefix(urlPath, prefixMust) || len(urlPath) < len(prefixMust)+2 {
		return "", "", "", fmt.Errorf("Request path must start with `%s`", prefixMust)
	}
	urlPath = urlPath[len(prefixMust):]
	if len(urlPath) < 3 {
		return "", "", "", fmt.Errorf("Unable to parse request from %q", urlPath)
	}

	parts := strings.SplitN(urlPath, "/", 4)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("Unable to parse request from %q", urlPath)
	}
	for i := len(parts); i < 4; i++ {
		parts = append(parts, "")
	}

	org := sanitize(parts[0])
	name := sanitize(strings.TrimSuffix(parts[1], ".git"))
	verb := parts[2]
	path := parts[3]

	return fmt.Sprintf("%s/%s", org, name), verb, path, nil
}

var alphaNum = regexp.MustCompile("[^a-z0-9-]+")

func sanitize(name string) string {
	return alphaNum.ReplaceAllString(strings.ToLower(name), "-")
}
