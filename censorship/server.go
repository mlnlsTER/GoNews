package main

import (
	"GoNews/censor/middleware"
	"log"
	"net/http"
	"strings"
)

func main() {
	censoredHandler := http.HandlerFunc(censorHandler)
	http.Handle("/", middleware.RequestIDMiddleware(middleware.LoggingMiddleware(censoredHandler)))

	log.Println("Censor service started on :8083...")
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatalf("Failed to start censor service: %v", err)
	}
}

func censorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	comment := r.FormValue("comment")
	if comment == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	if containsForbiddenWords(comment) {
		http.Error(w, "Comment contains forbidden words", http.StatusBadRequest)
		return
	}

	// Все проверки пройдены, комментарий прошел цензуру
	w.WriteHeader(http.StatusOK)
}

func containsForbiddenWords(content string) bool {
	forbiddenWords := []string{"qwerty", "йцукен", "zxvbnm"}
	for _, word := range forbiddenWords {
		if contentContainsWord(content, word) {
			return true
		}
	}
	return false
}

func contentContainsWord(content, word string) bool {
	return strings.Contains(strings.ToLower(content), strings.ToLower(word))
}
