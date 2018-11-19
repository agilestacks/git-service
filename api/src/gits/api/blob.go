package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"gits/config"
	"gits/repo"
)

func sendRepoBlob(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	path := filepath.Clean(vars["file"])

	ref := req.URL.Query().Get("ref")

	blobReader, err := repo.Blob(repoId, ref, path)
	if err != nil {
		message := fmt.Sprintf("Unable to obtain Git repo `%s` blob `%s`: %v", repoId, path, err)
		log.Print(message)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "is a directory") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not a regular file") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "file too big") {
			status = http.StatusNotImplemented
		}
		writeError(w, status, message)
	} else {
		defer blobReader.Close()
		if config.Verbose {
			log.Printf("Sending repo `%s` blob `%s`", repoId, path)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, blobReader)
	}
}
