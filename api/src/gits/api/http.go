package api

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	//"net/url"
	"strings"
	"time"

	"gits/config"
)

const (
	apiRepoHandlerPath = "/api/v1/repositories/"
)

func Listen(host string, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc(apiRepoHandlerPath, repositoriesRouter)
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
	json := fmt.Sprintf("{ \"error\": \"%s\" }", message)
	io.WriteString(w, json)
}

func repositoriesRouter(w http.ResponseWriter, req *http.Request) {
	if config.GitApiSecret != "" {
		secret := req.Header.Get("X-API-Secret")
		if secret != config.GitApiSecret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	repoId, verb, path, err := parsePath(req.URL.Path)
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
			log.Printf("HTTP %q request on repo %q; verb %q; path %q not allowed", req.Method, repoId, verb, path)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parsePath(urlPath string) (string, string, string, error) {
	if !strings.HasPrefix(urlPath, apiRepoHandlerPath) || len(urlPath) < len(apiRepoHandlerPath)+2 {
		return "", "", "", fmt.Errorf("Request path must start with `%s`", apiRepoHandlerPath)
	}
	urlPath = urlPath[len(apiRepoHandlerPath)+1:]
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

	return parts[0] + "/" + parts[1], parts[2], parts[3], nil
}

func sendRepoLog(repoId string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func createRepo(repoId string, bodyReader io.ReadCloser, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func uploadFile(repoId string, path string, req *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func uploadFiles(repoId string, req *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func deleteRepo(repoId string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
