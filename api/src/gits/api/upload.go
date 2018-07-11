package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"gits/config"
	"gits/repo"
)

func uploadFile(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	mode, err := queryFileMode(req)
	if err != nil {
		log.Printf("Bad file mode: %v", err)
	}
	add(repoId,
		[]repo.AddFile{{Path: vars["file"], Content: req.Body, Mode: mode}},
		queryCommitMessage(req),
		w)
}

func uploadFiles(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	diagnose := func(status int, message string) {
		log.Print(message)
		writeError(w, status, message)
	}

	err := req.ParseMultipartForm(config.MultipartMaxMemory)
	if err != nil {
		diagnose(http.StatusInternalServerError,
			fmt.Sprintf("Error parsing multipart request: %v", err))
		return
	}
	files := make([]repo.AddFile, 0, len(req.MultipartForm.File))
	for field, fileHeaders := range req.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			// file name is set to field name because Go parser strips directory from `filename`
			// what about `filepath`?
			name := field
			/*
				name := fileHeader.Filename
				if name == "" {
					diagnose(http.StatusBadRequest,
						fmt.Sprintf("Error processing multipart request part %q: filename is not set", field))
					return
				}
			*/
			file, err := fileHeader.Open()
			if err != nil {
				printFilename := ""
				if field != name {
					printFilename = fmt.Sprintf(" %q", name)
				}
				diagnose(http.StatusInternalServerError,
					fmt.Sprintf("Error processing multipart request part %q%s: %v", field, printFilename, err))
				return
			}
			defer file.Close()

			var mode os.FileMode
			if fileHeader.Header != nil {
				strModes, exist := fileHeader.Header["Mode"] // case-sensitive
				if exist && len(strModes) >= 1 && strModes[0] != "" {
					parsed, err := strconv.ParseUint(strModes[0], 8, 32)
					if err != nil {
						log.Printf("Bad file mode of %q %v: %v", name, strModes, err)
					} else {
						mode = os.FileMode(parsed)
					}
				}
			}

			files = append(files, repo.AddFile{Path: field, Content: file, Mode: mode})
			break
		}
	}
	add(repoId, files, queryCommitMessage(req), w)
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

func queryFileMode(req *http.Request) (os.FileMode, error) {
	strMode := req.URL.Query().Get("mode")
	mode, err := strconv.ParseUint(strMode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("Mode %q: %v", strMode, err)
	}
	return os.FileMode(mode), nil
}
