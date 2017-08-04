package api

import (
	"net/http"
	"time"
)

var hubApi = &http.Client{Timeout: 10 * time.Second}
var authApi = &http.Client{Timeout: 10 * time.Second}
