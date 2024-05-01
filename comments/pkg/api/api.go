package api

import (
	comStorage "GoNews/comments/pkg/storage"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type API struct {
	db comStorage.CommentsInterface
	r  *mux.Router
}

func (api *API) endpoints() {
	api.r.HandleFunc("/news/{newsID}", api.GetCommentsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/{newsID}", api.AddCommentHandler).Methods(http.MethodPost, http.MethodOptions)
}

func (api *API) Router() *mux.Router {
	return api.r
}

// Constructor creates a new API object.
func New(db comStorage.CommentsInterface) *API {
	api := API{
		db: db, r: mux.NewRouter(),
	}
	api.endpoints()
	return &api
}

func (api *API) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	newsID, err := strconv.ParseInt(mux.Vars(r)["newsID"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid news ID", http.StatusBadRequest)
		return
	}

	comments, err := api.db.Comments(newsID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func (api *API) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment comStorage.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := api.db.AddComments([]comStorage.Comment{comment}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
