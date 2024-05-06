package main

import (
	"GoNews/news/pkg/api"
	"GoNews/news/pkg/middleware"
	"GoNews/news/pkg/rss"
	newsStorage "GoNews/news/pkg/storage"
	"GoNews/news/pkg/storage/postgres"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// GoNews Server.
type server struct {
	db  newsStorage.NewsInterface
	api *api.API
}

type config struct {
	URLS   []string `json:"rss"`
	Period int      `json:"request_period"`
}

func main() {
	var err error
	var srv server

	// Initialize PostgreSQL server storage.
	srv.db, err = postgres.New("postgres://postgres:8952@localhost:5432/posts")
	if err != nil {
		log.Fatal(err)
	}

	// Create API object and register handlers.
	srv.api = api.New(srv.db)
	c, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	var conf config
	err = json.Unmarshal(c, &conf)
	if err != nil {
		log.Fatal(err)
	}

	chPosts := make(chan []newsStorage.Post)
	chErrors := make(chan error)

	for _, url := range conf.URLS {
		go parseURL(url, chPosts, chErrors, conf.Period)
	}

	go func() {
		for posts := range chPosts {
			srv.db.AddPosts(posts)
		}
	}()

	go func() {
		for err = range chErrors {
			log.Println(err)
		}
	}()
	log.Println("News service started on :8081...")
	err = http.ListenAndServe(":8081", middleware.RequestIDMiddleware(middleware.LoggingMiddleware(srv.api.Router())))
	if err != nil {
		log.Fatal(err)
	}
}

func parseURL(url string, chPosts chan<- []newsStorage.Post, chErrors chan<- error, peroid int) {
	for {
		posts, err := rss.Parse(url)
		if err != nil {
			chErrors <- err
			continue
		}
		chPosts <- posts
		time.Sleep(time.Duration(peroid) * time.Minute)
	}
}
