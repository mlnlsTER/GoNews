package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Middleware to add or retrieve a request identifier
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("request_id")
		if requestID == "" {
			requestID = uuid.NewString()[:8]
		}
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware for query logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value("request_id").(string)
		log.Printf("Request ID: %s | Time: %s | IP: %s\n", requestID, time.Now().Format(time.RFC3339), r.RemoteAddr)
		rw := &responseWriterWithStatus{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("Request ID: %s | Status Code: %d\n", requestID, rw.status)
	})
}

// responseWriterWithStatus - ResponseWriter proxy object with the ability to capture response status
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status of the response
func (rw *responseWriterWithStatus) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
