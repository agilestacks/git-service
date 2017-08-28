package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gits/config"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"
)

type middleware func(http.Handler) http.Handler

func mw(mws ...middleware) middleware {
	return func(handler http.Handler) http.Handler {
		h := handler
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}

func withLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		httptest.NewRecorder()

		log.Printf("%s: %s", req.Method, req.URL)
		crw := NewCapturingResponseWriter(rw, true)
		handler.ServeHTTP(crw, req)
		log.Printf("Response: %d %s", crw.Captured.Status, crw.Captured.Buffer.Bytes())
	})
}

func withApiSecret(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !checkApiSecret(req) {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(rw, req)
	})
}

func getRouter() http.Handler {
	r := mux.NewRouter()
	r.NotFoundHandler = mw(withLogger)(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))

	s := r.PathPrefix("/api/v1/repositories/{organization}/{repository}").Subrouter()
	s.Handle("", mw(withLogger, withApiSecret)(http.HandlerFunc(createRepo))).
		Methods("PUT")
	s.Handle("", mw(withLogger, withApiSecret)(http.HandlerFunc(deleteRepo))).
		Methods("DELETE")
	s.Handle("/commit/{file:.*}", mw(withLogger, withApiSecret)(http.HandlerFunc(uploadFile))).
		Methods("PUT")
	s.Handle("/commit", mw(withLogger, withApiSecret)(http.HandlerFunc(uploadFiles))).
		Methods("POST")
	s.Handle("/log", mw(withLogger, withApiSecret)(http.HandlerFunc(sendRepoLog))).
		Methods("GET")

	s = r.PathPrefix("/repo/{organization}/{repository}").Subrouter()
	s.Path("/info/refs").Queries("service", "git-upload-pack").
		Methods("GET").
		Handler(mw(withLogger, withApiSecret)(http.HandlerFunc(sendRefsInfo)))
	s.Path("/git-upload-pack").
		Methods("POST").
		Handler(mw(withLogger, withApiSecret)(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Content-Encoding") == "gzip" {
				in, err := gzip.NewReader(req.Body)
				if err != nil {
					log.Printf("Unable to decompress request: %v", err)
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}
				req.Body = ioutil.NewReadCloser(in, req.Body)
			}

			sendRefsPack(rw, req)
		})))

	return r
}

func Listen(host string, port int) {
	r := getRouter()

	http.Handle("/", r)

	go listen(&http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})
}

func listen(server *http.Server) {
	log.Fatalf("Error in HTTP server: %v", server.ListenAndServe())
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{message})

	if err != nil {
		log.Println("unable to marshall json", err)
		b = []byte(err.Error())
	}

	w.Write(b)
}

func checkApiSecret(req *http.Request) bool {
	if config.GitApiSecret == "" {
		return true
	}

	xApiSecret := req.Header.Get("X-API-Secret")
	if config.Trace {
		log.Printf("X-API-Secret `%v`", xApiSecret)
	}
	if xApiSecret == config.GitApiSecret {
		return true
	}

	// for git clone https://
	user, password, ok := req.BasicAuth()
	if config.Trace {
		log.Printf("Authorization `%v`; Basic auth %v `%v` `%v`",
			req.Header.Get("Authorization"), ok, user, password)
	}
	if ok {
		return user == config.GitApiSecret || password == config.GitApiSecret
	}

	return false
}

var alphaNum = regexp.MustCompile("[^a-z0-9-]+")

func sanitize(name string) string {
	return alphaNum.ReplaceAllString(strings.ToLower(name), "-")
}

func getRepositoryId(org string, repo string) string {
	return fmt.Sprintf("%s/%s",
		sanitize(org),
		sanitize(strings.TrimSuffix(repo, ".git")))
}
