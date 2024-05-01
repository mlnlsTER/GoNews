package memdb

import (
	"GoNews/comments/pkg/storage"
)

// Хранилище данных.
type DB []comStorage.Comment

func (db *DB) Comments() ([]comStorage.Comment, error) {
	return *db, nil
}
func (db *DB) AddComments(comment comStorage.Comment) error {
	*db = append(*db, comment)
	return nil
}
