package postgres

import (
	"context"
	"GoNews/comments/pkg/storage"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Data storage.
type Storage struct {
	db *pgxpool.Pool
}

// Constructor creates a new Storage object.
func New(constr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Storage{
		db: db,
	}
	return &s, nil
}

// Comments returns publication comments from the database.
func (s *Storage) Comments(n int64) ([]comStorage.Comment, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			id,
			id_news,
			id_parent,
			content,
			commented_at
		FROM comments
		WHERE id_news = $1
		ORDER BY id DESC;
	`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []comStorage.Comment
	for rows.Next() {
		var c comStorage.Comment
		err = rows.Scan(
			&c.ID,
			&c.ID_News,
			&c.ID_Parent,
			&c.Content,
			&c.ComTime,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)

	}
	return comments, rows.Err()
}

// AddComments creates a new comments in the database.
func (s *Storage) AddComments(comments []comStorage.Comment) error {
	for _, comment := range comments {
		_, err := s.db.Exec(context.Background(), `
		INSERT INTO comments(id, id_news, id_parent, content, commented_at)
		VALUES ($1, $2, $3, $4, $5)`,
			comment.ID,
			comment.ID_News,
			comment.ID_Parent,
			comment.Content,
			comment.ComTime,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
