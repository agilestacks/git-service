package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gits/config"
	"gits/repo"
)

func uploadFile(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	add(repoId,
		[]repo.AddFile{{Path: vars["file"], Content: req.Body}},
		queryCommitMessage(req),
		w)
}

func uploadFiles(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	err := req.ParseMultipartForm(config.MultipartMaxMemory)
	if err != nil {
		message := fmt.Sprintf("Error parsing multipart request: %v", err)
		log.Print(message)
		writeError(w, http.StatusBadRequest, message)
	} else {
		files := make([]repo.AddFile, 0, len(req.MultipartForm.File))
		for name, fileHeader := range req.MultipartForm.File {
			file, err := fileHeader[0].Open()
			if err != nil {
				message := fmt.Sprintf("Error processing multipart request part %q: %v", name, err)
				log.Print(message)
				writeError(w, http.StatusInternalServerError, message)
				return
			}
			defer file.Close()
			files = append(files, repo.AddFile{Path: name, Content: file})
		}
		add(repoId, files, queryCommitMessage(req), w)
	}
}

func add(repoId string, files []repo.AddFile, commitMessage string, w http.ResponseWriter) {
	err := repo.Add(repoId, files, commitMessage)
	if err != nil {
		message := fmt.Sprintf("Unable to add files to Git repository: %v", err)
		log.Print(message)
		writeError(w, http.StatusInternalServerError, message)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func queryCommitMessage(req *http.Request) string {
	return req.URL.Query().Get("message")
}
