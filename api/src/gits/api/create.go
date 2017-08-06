package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"gits/config"
	"gits/repo"
)

type CreateRequest struct {
	Archive string
}

func createRepo(repoId string, bodyReader io.Reader, w http.ResponseWriter) {
	var body bytes.Buffer
	read, err := body.ReadFrom(bodyReader)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading request body (read %d bytes): %v", read, err))
		return
	}
	archive := ""
	if read > 4 {
		var req CreateRequest
		err = json.Unmarshal(body.Bytes(), &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Error unmarshalling JSON request: %v", err))
			return
		}
		archive = req.Archive
	}

	err = repo.Create(repoId, archive)
	if err != nil {
		message := fmt.Sprintf("Unable to create Git repo `%s`: %v", repoId, err)
		log.Print(message)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exist") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "not supported") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "S3") {
			status = http.StatusGatewayTimeout
		}
		writeError(w, status, message)
	} else {
		if config.Verbose {
			log.Printf("Repo `%s` created", repoId)
		}
		w.WriteHeader(http.StatusCreated)
	}
}
