package posts

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrTitleRequired = errors.New("post title required")
	ErrBodyRequired  = errors.New("post body required")
)

type Service struct {
	repo Repository
}

const (
	defaultPostsLimit = 20
	maxPostsLimit     = 100
)

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListWithPagination(ctx context.Context, limit, offset int) ([]Post, error) {
	normalizedLimit := normalizePostsLimit(limit)
	normalizedOffset := normalizePostsOffset(offset)
	posts, err := s.repo.ListPosts(ctx, normalizedLimit, normalizedOffset)
	if err != nil {
		return nil, err
	}

	payload := make([]Post, 0, len(posts))
	for _, post := range posts {
		payload = append(payload, mapStoredPost(post))
	}
	return payload, nil
}

func (s *Service) Create(ctx context.Context, authorUserID, title, body string) (Post, error) {
	normalizedTitle := strings.TrimSpace(title)
	if normalizedTitle == "" {
		return Post{}, ErrTitleRequired
	}

	normalizedBody := strings.TrimSpace(body)
	if normalizedBody == "" {
		return Post{}, ErrBodyRequired
	}

	post, err := s.repo.CreatePost(ctx, authorUserID, normalizedTitle, normalizedBody)
	if err != nil {
		return Post{}, err
	}

	return mapStoredPost(post), nil
}

func (s *Service) GetByID(ctx context.Context, postID string) (Post, error) {
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			return Post{}, ErrPostNotFound
		}
		return Post{}, err
	}
	return mapStoredPost(post), nil
}

func mapStoredPost(post StoredPost) Post {
	return Post{
		ID:           post.ID,
		AuthorUserID: post.AuthorUserID,
		Title:        post.Title,
		Body:         post.Body,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}

func normalizePostsLimit(limit int) int {
	if limit <= 0 {
		return defaultPostsLimit
	}
	if limit > maxPostsLimit {
		return maxPostsLimit
	}
	return limit
}

func normalizePostsOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
