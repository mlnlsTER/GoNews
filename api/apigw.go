package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"GoNews/api/middleware"

	"github.com/gorilla/mux"
)

// RequestResult stores the result of the request and the error.
type RequestResult struct {
	Data interface{}
	Err  error
}

// Get the list of news.
type NewsListResponse struct {
	Posts []NewsShortDetailed `json:"posts"`
}

// NewsFullDetailed contains complete information about a news item.
type NewsFullDetailed struct {
	ID       int
	Title    string
	Content  string
	PubTime  int64
	Link     string
	Comments []Comment
}

// NewsShortDetailed contains brief information about the news.
type NewsShortDetailed struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

// Comment contains information about the comment.
type Comment struct {
	ID        int    // comment number
	ID_News   int64  // news number
	ID_Parent int64  // parent number (if the answer to the comment)
	Content   string // comment content
	ComTime   int64  // comment time
}

// AddCommentHandler handles a request to add a comment to a news item.
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	newsID, err := strconv.Atoi(vars["newsID"])
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	comment.ID_News = int64(newsID)

	resp, err := http.Post("http://localhost:8083/", "application/x-www-form-urlencoded", strings.NewReader("comment="+comment.Content))
	if err != nil {
		http.Error(w, "Failed to check comment censorship", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Comment contains forbidden words", http.StatusBadRequest)
		return
	}

	commentJSON, err := json.Marshal(comment)
	if err != nil {
		http.Error(w, "Failed to marshal comment", http.StatusInternalServerError)
		return
	}

	resp, err = http.Post("http://localhost:8082/news/"+vars["newsID"], "application/json", bytes.NewBuffer(commentJSON))
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		http.Error(w, "Failed to add comment", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetNewsDetailHandler processes a request to get detailed information about a news item.
func GetNewsDetailHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	newsID := vars["newsID"]
	chResults := make(chan RequestResult, 2)
	var wg sync.WaitGroup

	// Запрос к агрегатору новостей
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := http.Get("http://localhost:8081/news/" + newsID)
		if err != nil {
			chResults <- RequestResult{Err: err}
			return
		}
		defer resp.Body.Close()

		var newsDetail NewsFullDetailed
		if err := json.NewDecoder(resp.Body).Decode(&newsDetail); err != nil {
			chResults <- RequestResult{Err: err}
			return
		}

		chResults <- RequestResult{Data: newsDetail}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		client := http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:8082/news/"+newsID, nil)
		if err != nil {
			chResults <- RequestResult{Err: err}
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			chResults <- RequestResult{Err: err}
			return
		}
		defer resp.Body.Close()

		var comments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			chResults <- RequestResult{Err: err}
			return
		}

		chResults <- RequestResult{Data: comments}
	}()
	go func() {
		wg.Wait()
		close(chResults)
	}()
	var newsData NewsFullDetailed
	var commentsData []Comment

	for result := range chResults {
		if result.Err != nil {
			http.Error(w, result.Err.Error(), http.StatusInternalServerError)
			return
		}

		switch data := result.Data.(type) {
		case NewsFullDetailed:
			newsData = data
		case []Comment:
			commentsData = data
		}
	}
	newsData.Comments = commentsData
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(newsData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetNewsListHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	resp, err := http.Get("http://localhost:8081/news?page=" + pageStr)
	if err != nil {
		http.Error(w, "Failed to fetch news list", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch news list", resp.StatusCode)
		return
	}
	var newsListResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsListResponse); err != nil {
		http.Error(w, "Failed to decode news list response", http.StatusInternalServerError)
		return
	}

	var shortNewsList []NewsShortDetailed
	for _, news := range newsListResponse.Posts {
		var shortContent string
		if len(news.Content) > 200 {
			shortContent = news.Content[:200] + "..."
		} else {
			shortContent = news.Content
		}

		shortNews := NewsShortDetailed{
			ID:      news.ID,
			Title:   news.Title,
			Content: shortContent,
			PubTime: news.PubTime,
			Link:    news.Link,
		}
		shortNewsList = append(shortNewsList, shortNews)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shortNewsList); err != nil {
		http.Error(w, "Failed to encode news list", http.StatusInternalServerError)
		return
	}
}

// FilterNewsHandler handles the request to filter the news list.
func FilterNewsHandler(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	searchParam := url.QueryEscape(r.URL.Query().Get("s"))
	pageStr := r.URL.Query().Get("page")
	log.Println(searchParam)
	if pageStr != "" {
		resp, err = http.Get("http://localhost:8081/news?s=" + searchParam + "&page=" + pageStr)
	} else {
		resp, err = http.Get("http://localhost:8081/news?s=" + searchParam)
	}
	if err != nil {
		http.Error(w, "Failed to fetch filtered news list", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch filtered news list", resp.StatusCode)
		return
	}

	var filteredNewsListResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&filteredNewsListResponse); err != nil {
		http.Error(w, "Failed to decode filtered news list", http.StatusInternalServerError)
		return
	}

	var filteredNewsList []NewsShortDetailed
	for _, news := range filteredNewsListResponse.Posts {
		var shortContent string
		if len(news.Content) > 200 {
			shortContent = news.Content[:200] + "..."
		} else {
			shortContent = news.Content
		}

		shortNews := NewsShortDetailed{
			ID:      news.ID,
			Title:   news.Title,
			Content: shortContent,
			PubTime: news.PubTime,
			Link:    news.Link,
		}
		filteredNewsList = append(filteredNewsList, shortNews)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredNewsList); err != nil {
		http.Error(w, "Failed to encode news list", http.StatusInternalServerError)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/news/{newsID:[0-9]+}", AddCommentHandler).Methods("POST")
	router.HandleFunc("/news/{newsID:[0-9]+}", GetNewsDetailHandler).Methods("GET")
	router.HandleFunc("/news/", GetNewsListHandler).Methods("GET")
	router.HandleFunc("/news/filter", FilterNewsHandler).Methods("GET")

	log.Println("API Gateway запущен на порту 8080...")
	err := http.ListenAndServe(":8080", middleware.RequestIDMiddleware(middleware.LoggingMiddleware(router)))
	log.Fatal(err)
}
