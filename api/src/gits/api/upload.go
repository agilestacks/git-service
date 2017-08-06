package api

import (
	"net/http"
)

func uploadFile(repoId string, path string, req *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func uploadFiles(repoId string, req *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
