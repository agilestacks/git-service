package api

import (
	"net/http"
)

func deleteRepo(repoId string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
