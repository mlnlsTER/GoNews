package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"GoNews/api/middleware"

	"github.com/gorilla/mux"
)

// RequestResult хранит результат запроса и ошибку.
type RequestResult struct {
	Data interface{}
	Err  error
}

// Получение списка новостей.
type NewsListResponse struct {
	Posts []NewsShortDetailed `json:"posts"`
}

// NewsFullDetailed содержит полную информацию о новости.
type NewsFullDetailed struct {
	ID       int
	Title    string
	Content  string
	PubTime  int64
	Link     string
	Comments []Comment
}

// NewsShortDetailed содержит краткую информацию о новости.
type NewsShortDetailed struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

// Comment содержит информацию о комментарии.
type Comment struct {
	ID        int    // comment number
	ID_News   int64  // news number
	ID_Parent int64  // parent number (if the answer to the comment)
	Content   string // comment content
	ComTime   int64  // comment time
}

// AddCommentHandler обрабатывает запрос на добавление комментария к новости.
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

	// Проверяем комментарий на цензуру
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

	// Преобразуем комментарий в JSON
	commentJSON, err := json.Marshal(comment)
	if err != nil {
		http.Error(w, "Failed to marshal comment", http.StatusInternalServerError)
		return
	}

	// Если комментарий прошел проверку на цензуру, добавляем его через API сервиса комментариев
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

// GetNewsDetailHandler обрабатывает запрос на получение детальной информации о новости.
func GetNewsDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор новости из URL
	vars := mux.Vars(r)
	newsID := vars["newsID"]
	// Создаем канал для передачи результатов запросов и ошибок
	chResults := make(chan RequestResult, 2)

	// Создаем группу ожидания
	var wg sync.WaitGroup

	// Запускаем горутины для выполнения запросов

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

	// Запрос к сервису комментариев
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		client := http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:8082/news/"+newsID, nil)
		if err != nil {
			chResults <- RequestResult{Err: err}
			return
		}
		req.Header.Set("Content-Type", "application/json") // Установка заголовка Content-Type
		
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

	// Ожидаем завершения всех запросов
	go func() {
		wg.Wait()
		close(chResults)
	}()

	// Обрабатываем результаты запросов

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

	// Добавляем комментарии к новости
	newsData.Comments = commentsData

	// Возвращаем результат в формате JSON
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

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch news list", resp.StatusCode)
		return
	}

	// Декодируем полученные данные из JSON
	var newsListResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsListResponse); err != nil {
		http.Error(w, "Failed to decode news list response", http.StatusInternalServerError)
		return
	}

	// Преобразуем полученные данные в формат NewsShortDetailed
	var shortNewsList []NewsShortDetailed
	// Преобразование данных из NewsFullDetailed в NewsShortDetailed с обрезанием Content до 200 символов
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

	// Возвращаем список новостей в формате JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shortNewsList); err != nil {
		http.Error(w, "Failed to encode news list", http.StatusInternalServerError)
		return
	}
}

// FilterNewsHandler обрабатывает запрос на фильтрацию списка новостей.
func FilterNewsHandler(w http.ResponseWriter, r *http.Request) {
	searchParam := r.URL.Query().Get("s")
	fmt.Println(url.QueryEscape(r.URL.Query().Get("s")))
	resp, err := http.Get("http://localhost:8081/news?s=" + searchParam)
	if err != nil {
		http.Error(w, "Failed to fetch filtered news list", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch filtered news list", resp.StatusCode)
		return
	}

	// Декодируем полученные данные из JSON
	var filteredNewsListResponse NewsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&filteredNewsListResponse); err != nil {
		http.Error(w, "Failed to decode filtered news list", http.StatusInternalServerError)
		return
	}
	// Преобразуем полученные данные в формат NewsShortDetailed
	var filteredNewsList []NewsShortDetailed
	// Преобразование данных из NewsFullDetailed в NewsShortDetailed с обрезанием Content до 200 символов
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

	// Возвращаем список новостей в формате JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredNewsList); err != nil {
		http.Error(w, "Failed to encode news list", http.StatusInternalServerError)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/news/{newsID}", AddCommentHandler).Methods("POST")
	router.HandleFunc("/news/{newsID}", GetNewsDetailHandler).Methods("GET")
	router.HandleFunc("/news", GetNewsListHandler).Methods("GET")
	router.HandleFunc("/news", FilterNewsHandler).Methods("GET")

	log.Println("API Gateway запущен на порту 8080...")
	err := http.ListenAndServe(":8080", middleware.RequestIDMiddleware(middleware.LoggingMiddleware(router)))
	log.Fatal(err)
}
