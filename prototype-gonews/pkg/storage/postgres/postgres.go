package postgres

import (
	newsStorage "GoNews/news/pkg/storage"
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const pageSize int = 15

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

// PostDetail returns detailed information about a publication by its identifier.
func (s *Storage) PostDetail(id int) (*newsStorage.Post, error) {
	var post newsStorage.Post

	err := s.db.QueryRow(context.Background(), `
		SELECT 
			id,
			title,
			content,
			published_at,
			link
		FROM posts
		WHERE id = $1
	`, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.PubTime,
		&post.Link,
	)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// Posts returns the publications from the database.
func (s *Storage) Posts(page int, searchQuery string) ([]newsStorage.Post, newsStorage.Pagination, error) {
	if page <= 0 {
		page = 1
	}
	var totalPages int
	var totalPosts int
	var rows pgx.Rows
	var err error
	if searchQuery != "" {
		err = s.db.QueryRow(context.Background(), `
            SELECT COUNT(*)
            FROM posts
            WHERE title ILIKE $1;
        `, "%"+searchQuery+"%").Scan(&totalPosts)
		if err != nil {
			return nil, newsStorage.Pagination{}, err
		}
		totalPages = (totalPosts + pageSize - 1) / pageSize
		offset := (page - 1) * pageSize
		rows, err = s.db.Query(context.Background(), `
            SELECT 
                id,
                title,
                content,
                published_at,
                link
            FROM posts
            WHERE title ILIKE $1
            ORDER BY id DESC
            LIMIT $2 OFFSET $3;
        `, "%"+searchQuery+"%", pageSize, offset)
	} else {
		err = s.db.QueryRow(context.Background(), `
            SELECT COUNT(*)
            FROM posts;
        `).Scan(&totalPosts)
		if err != nil {
			return nil, newsStorage.Pagination{}, err
		}
		totalPages = (totalPosts + pageSize - 1) / pageSize
		offset := (page - 1) * pageSize
		rows, err = s.db.Query(context.Background(), `
            SELECT 
                id,
                title,
                content,
                published_at,
                link
            FROM posts
            ORDER BY id DESC
            LIMIT $1 OFFSET $2;
        `, pageSize, offset)
	}

	if err != nil {
		return nil, newsStorage.Pagination{}, err
	}
	defer rows.Close()

	var posts []newsStorage.Post
	for rows.Next() {
		var p newsStorage.Post
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.PubTime,
			&p.Link,
		)
		if err != nil {
			return nil, newsStorage.Pagination{}, err
		}
		posts = append(posts, p)
	}

	pagination := newsStorage.Pagination{
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}

	return posts, pagination, nil
}

// AddPosts creates a new publications in the database.
func (s *Storage) AddPosts(posts []newsStorage.Post) error {
	for _, post := range posts {
		_, err := s.db.Exec(context.Background(), `
		INSERT INTO posts(title, content, published_at, link)
		VALUES ($1, $2, $3, $4)`,
			post.Title,
			post.Content,
			post.PubTime,
			post.Link,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
