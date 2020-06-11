package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/util"
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

		if config.Debug {
			log.Printf("HTTP <<< %s %s", req.Method, req.URL)
		}
		crw := NewCapturingResponseWriter(rw, false)
		handler.ServeHTTP(crw, req)
		if config.Debug {
			log.Printf("HTTP === %d", crw.Captured.Status)
		}
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

func withAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !checkApiSecretOrUserAuth(req) {
			rw.Header().Set("WWW-Authenticate", "Basic realm=\".\"")
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(rw, req)
	})
}

func withRepoExist(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !checkRepoExist(req) {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		handler.ServeHTTP(rw, req)
	})
}

func withAllowedGitService(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !checkGitService(req) {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		handler.ServeHTTP(rw, req)
	})
}

func gunzip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Encoding") == "gzip" {
			in, err := gzip.NewReader(req.Body)
			if err != nil {
				log.Printf("Unable to decompress request: %v", err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = util.NewReadCloser(in, req.Body)
		}

		handler.ServeHTTP(rw, req)
	})
}

func rejectIfMaintenance(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if maint, msg := util.Maintenance(); maint {
			rw.WriteHeader(http.StatusServiceUnavailable)
			if msg != "" {
				rw.Write([]byte(msg))
			}
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
	cmw := mw(withLogger, withApiSecret, withRepoExist)
	s.Handle("", mw(withLogger, withApiSecret, rejectIfMaintenance)(http.HandlerFunc(createRepo))).
		Methods("PUT")
	s.Handle("", mw(cmw, rejectIfMaintenance)(http.HandlerFunc(deleteRepo))).
		Methods("DELETE")
	s.Handle("/commit/{file:.*}", mw(cmw, rejectIfMaintenance)(http.HandlerFunc(uploadFile))).
		Methods("PUT")
	s.Handle("/commit", mw(cmw, rejectIfMaintenance)(http.HandlerFunc(uploadFiles))).
		Methods("POST")
	s.Handle("/subtrees", mw(cmw, rejectIfMaintenance)(http.HandlerFunc(addSubtrees))).
		Methods("POST")
	s.Handle("/blob/{file:.*}", cmw(http.HandlerFunc(sendRepoBlob))).
		Methods("GET")
	s.Handle("/log", cmw(http.HandlerFunc(sendRepoLog))).
		Methods("GET")
	s.Handle("/status", cmw(http.HandlerFunc(sendRepoStatus))).
		Methods("GET")

	s = r.PathPrefix("/repo/{organization}/{repository}").Subrouter()
	cmw = mw(withLogger, withAuth, withAllowedGitService, withRepoExist)
	s.Path("/info/refs").Queries("service", "{service}").
		Methods("GET").
		Handler(cmw(http.HandlerFunc(refsInfo)))
	s.Path("/{service}").
		Methods("POST").
		Handler(mw(cmw, rejectIfMaintenance, gunzip)(http.HandlerFunc(pack)))

	s = r.PathPrefix("/api/v1/ping").Subrouter()
	s.Handle("", mw(withLogger)(http.HandlerFunc(ping))).
		Methods("GET")

	return r
}

func ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func Listen(host string, port int) {
	r := getRouter()

	http.Handle("/", r)

	go listen(&http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
	})
}

func listen(server *http.Server) {
	log.Fatalf("Error in HTTP server: %v", server.ListenAndServe())
}

func writeError(w http.ResponseWriter, status int, message string) {
	if config.Debug {
		log.Printf("Error %d HTTP: %s", status, message)
	}

	b, err := json.Marshal(struct {
		Error string `json:"error"`
	}{message})

	if err != nil {
		msg := fmt.Sprintf("Unable to marshall JSON: %v", err)
		log.Print(msg)
		b = []byte(msg)
		w.Header().Set("Content-Type", "text/plain")
		status = http.StatusInternalServerError
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(status)
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
	username, password, ok := req.BasicAuth()
	if config.Trace {
		log.Printf("Authorization `%v`; Basic auth %v `%v` `%v`",
			req.Header.Get("Authorization"), ok, username, password)
	}
	if ok {
		return username == config.GitApiSecret || password == config.GitApiSecret
	}

	return false
}

func checkApiSecretOrUserAuth(req *http.Request) bool {
	if checkApiSecret(req) {
		return true
	}
	return checkUserRepoAccess(req)
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
