package memdb

import (
	"GoNews/news/pkg/storage"
)

// Data storage.
type DB []newsStorage.Post

func (db *DB) Posts() ([]newsStorage.Post, error) {
	return *db, nil
}
func (db *DB) AddPost(post newsStorage.Post) error {
	*db = append(*db, post)
	return nil
}
