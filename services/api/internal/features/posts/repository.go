package posts

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

type StoredPost struct {
	ID           string
	AuthorUserID string
	Title        string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	ListPosts(ctx context.Context, limit, offset int) ([]StoredPost, error)
	CreatePost(ctx context.Context, authorUserID, title, body string) (StoredPost, error)
	GetPostByID(ctx context.Context, postID string) (StoredPost, error)
}

type PGRepository struct {
	dbPool *pgxpool.Pool
}

func NewPGRepository(dbPool *pgxpool.Pool) *PGRepository {
	return &PGRepository{dbPool: dbPool}
}

func (r *PGRepository) ListPosts(ctx context.Context, limit, offset int) ([]StoredPost, error) {
	rows, err := r.dbPool.Query(ctx, `
		SELECT id, author_user_id, title, body, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]StoredPost, 0)
	for rows.Next() {
		var post StoredPost
		if err := rows.Scan(&post.ID, &post.AuthorUserID, &post.Title, &post.Body, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return posts, nil
}

func (r *PGRepository) CreatePost(ctx context.Context, authorUserID, title, body string) (StoredPost, error) {
	var post StoredPost
	err := r.dbPool.QueryRow(ctx, `
		INSERT INTO posts (author_user_id, title, body)
		VALUES ($1, $2, $3)
		RETURNING id, author_user_id, title, body, created_at, updated_at
	`, authorUserID, title, body).Scan(
		&post.ID,
		&post.AuthorUserID,
		&post.Title,
		&post.Body,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return StoredPost{}, err
	}
	return post, nil
}

func (r *PGRepository) GetPostByID(ctx context.Context, postID string) (StoredPost, error) {
	var post StoredPost
	err := r.dbPool.QueryRow(ctx, `
		SELECT id, author_user_id, title, body, created_at, updated_at
		FROM posts
		WHERE id = $1
	`, postID).Scan(
		&post.ID,
		&post.AuthorUserID,
		&post.Title,
		&post.Body,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredPost{}, ErrPostNotFound
		}
		return StoredPost{}, err
	}
	return post, nil
}
