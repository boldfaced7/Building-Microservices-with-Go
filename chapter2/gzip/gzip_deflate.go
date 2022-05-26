package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var name string

type HelloWorldResponse struct {
	Message string `json:"message"`
}

func main() {
	port := 8080

	http.Handle("/helloworld",
		NewGzipHandler(http.HandlerFunc(helloWorldHandler)),
	)

	log.Printf("Server starting on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func helloWorldHandler(rw http.ResponseWriter, r *http.Request) {
	response := HelloWorldResponse{Message: "Hello " + name}

	encoder := json.NewEncoder(rw)
	encoder.Encode(response)
}

func NewGzipHandler(next http.Handler) http.Handler {
	return &GzipHandler{next}
}

type GzipHandler struct {
	next http.Handler
}

func (h *GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	encodings := r.Header.Get("Accept-Encoding")

	if strings.Contains(encodings, "gzip") {
		h.serveGzipped(w, r)
	} else if strings.Contains(encodings, "deflate") {
		panic("Deflate not implemented")
	} else {
		h.ServePlain(w, r)
	}
}

func (h *GzipHandler) serveGzipped(w http.ResponseWriter, r *http.Request) {
	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	w.Header().Set("Content-Encoding", "gzip")
	h.next.ServeHTTP(GzipResponseWriter{gzw, w}, r)
}

func (h *GzipHandler) ServePlain(w http.ResponseWriter, r *http.Request) {
	h.next.ServeHTTP(w, r)
}

type GzipResponseWriter struct {
	gw *gzip.Writer
	http.ResponseWriter
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
	if _, ok := w.Header()["Content-Type"]; !ok {
		// 콘텐츠 타입이 설정되지 않은 경우, 압축되지 않은 본문에서 콘텐츠 타입을 유추
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.gw.Write(b)
}

func (w GzipResponseWriter) Flush() {
	w.gw.Flush()
	if fw, ok := w.ResponseWriter.(http.Flusher); ok {
		fw.Flush()
	}
}
