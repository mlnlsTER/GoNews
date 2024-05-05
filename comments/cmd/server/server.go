package main

import (
	"log"
	"net/http"
	"strings"
	"GoNews/comments/pkg/api"
	"GoNews/comments/pkg/storage"
	"GoNews/comments/pkg/storage/postgres"
	"GoNews/comments/pkg/middleware"
)

type server struct {
	db  comStorage.CommentsInterface
	api *api.API
}

func main() {
	var err error
	var srv server

	srv.db, err = postgres.New("postgres://postgres:8952@localhost:5432/comments")
	if err != nil {
		log.Fatal(err)
	}
	srv.api = api.New(srv.db)

	chComments := make(chan []comStorage.Comment)
	chErrors := make(chan error)

	go func() {
		for comments := range chComments {
			censorComments(comments, srv.db, chErrors)
		}
	}()

	go func() {
		for err = range chErrors {
			log.Println(err)
		}
	}()

	log.Println("Comments service started on :8082...")
	err = http.ListenAndServe(":8082", middleware.RequestIDMiddleware(middleware.LoggingMiddleware(srv.api.Router())))
	if err != nil {
		log.Fatal(err)
	}
}


// censorComments censors comments for inappropriate content
func censorComments(comments []comStorage.Comment, db comStorage.CommentsInterface, chErrors chan<- error) {
	approvedComments := make([]comStorage.Comment, 0)
	for _, comment := range comments {
		if !containsForbiddenWords(comment.Content) {
			approvedComments = append(approvedComments, comment)
		} else {
			log.Println("Comment contains forbidden words and is blocked:", comment)
		}
	}

	err := db.AddComments(approvedComments)
	if err != nil {
		chErrors <- err
	}
}

// containsForbiddenWords checks if a comment contains forbidden words
func containsForbiddenWords(content string) bool {
	forbiddenWords := []string{"qwerty", "йцукен", "zxvbnm"}
	for _, word := range forbiddenWords {
		if contentContainsWord(content, word) {
			return true
		}
	}
	return false
}

// contentContainsWord checks if a string contains a word
func contentContainsWord(content, word string) bool {
	return strings.Contains(strings.ToLower(content), strings.ToLower(word))
}
