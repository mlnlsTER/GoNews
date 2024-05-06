package api

import (
	newsStorage "GoNews/news/pkg/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API - API object.
type API struct {
	db     newsStorage.NewsInterface
	router *mux.Router
}

// Constructor creates a new API object.
func New(db newsStorage.NewsInterface) *API {
	api := API{
		db: db, router: mux.NewRouter(),
	}
	api.endpoints()
	return &api
}

// Registering API endpoints.
func (api *API) endpoints() {
	api.router.HandleFunc("/news", api.PostsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/news/{id}", api.PostDetailHandler).Methods(http.MethodGet, http.MethodOptions)

	api.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("cmd/server/webapp"))))
}

// Receive request router.
func (api *API) Router() *mux.Router {
	return api.router
}

// Retrieving a list of news items.
func (api *API) PostsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var page int
	var pagination newsStorage.Pagination
	var posts []newsStorage.Post
	searchStr := r.URL.Query().Get("s")
	fmt.Println(searchStr)
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
	} else {
		page = 1
	}
	if err != nil {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}
	posts, pagination, err = api.db.Posts(page, searchStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseData := struct {
		Posts      []newsStorage.Post     `json:"posts"`
		Pagination newsStorage.Pagination `json:"pagination"`
	}{
		Posts:      posts,
		Pagination: pagination,
	}
	json.NewEncoder(w).Encode(responseData)
}

// Getting detailed information about the news.
func (api *API) PostDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	post, err := api.db.PostDetail(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(post)
}
