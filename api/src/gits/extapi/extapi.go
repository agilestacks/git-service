package extapi

import (
	"net/http"
	"time"
)

var hubApi = &http.Client{Timeout: 10 * time.Second}
var authApi = &http.Client{Timeout: 10 * time.Second}
var subsApi = &http.Client{Timeout: 10 * time.Second}
