package api

import (
	"net/http"
)

func sendRepoLog(repoId string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
