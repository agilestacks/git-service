package api

import (
	"bytes"
	"net/http"
)

type capturingWriter struct {
	http.ResponseWriter
	Captured CapturedResponse
}

type CapturedResponse struct {
	Status int
	Header http.Header
	Buffer *bytes.Buffer
}

// NewCaptureWriter wraps an existing ResponseWrapper into a proxy
// and caches response body and status written to it
func NewCapturingResponseWriter(w http.ResponseWriter, captureBody bool) *capturingWriter {
	cw := &capturingWriter{
		ResponseWriter: w,
	}

	cw.Captured.Status = 200
	if captureBody {
		cw.Captured.Buffer = new(bytes.Buffer)
	}

	return cw
}

func (cw *capturingWriter) WriteHeader(status int) {
	cw.Captured.Status = status
	cw.copyHeaders()
	cw.ResponseWriter.WriteHeader(status)
}

func (cw *capturingWriter) Write(bytes []byte) (int, error) {
	cw.copyHeaders()
	size, err := cw.ResponseWriter.Write(bytes)
	if err == nil && cw.Captured.Buffer != nil {
		cw.Captured.Buffer.Write(bytes)
	}
	return size, err
}

func (cw *capturingWriter) copyHeaders() {
	if cw.Captured.Header != nil {
		return
	}

	cw.Captured.Header = make(http.Header)
	for k, v := range cw.ResponseWriter.Header() {
		cw.Captured.Header[k] = v
	}
}
