package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Middleware для добавления или извлечения идентификатора запроса
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

// Middleware для журналирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получение идентификатора запроса из контекста
		requestID := r.Context().Value("request_id").(string)

		// Журналирование времени запроса и IP-адреса отправителя
		log.Printf("Request ID: %s | Time: %s | IP: %s\n", requestID, time.Now().Format(time.RFC3339), r.RemoteAddr)

		// Создание прокси-объекта ResponseWriter с перехватом статуса ответа
		rw := &responseWriterWithStatus{ResponseWriter: w, status: http.StatusOK}

		// Переход к обработчику запроса с прокси-объектом ResponseWriter
		next.ServeHTTP(rw, r)

		// Журналирование HTTP-кода ответа
		log.Printf("Request ID: %s | Status Code: %d\n", requestID, rw.status)
	})
}

// responseWriterWithStatus - прокси-объект ResponseWriter с возможностью перехвата статуса ответа
type responseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

// WriteHeader перехватывает статус ответа
func (rw *responseWriterWithStatus) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
