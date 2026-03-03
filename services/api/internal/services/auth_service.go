package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"example.com/api/internal/auth"
	"example.com/api/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type AuthService struct {
	repo      repositories.AuthRepository
	jwtSecret []byte
}

func NewAuthService(repo repositories.AuthRepository, jwtSecret []byte) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", User{}, err
	}

	created, err := s.repo.CreateUser(ctx, normalizedEmail, string(hash))
	if err != nil {
		if errors.Is(err, repositories.ErrAuthEmailExists) {
			return "", User{}, ErrEmailExists
		}
		return "", User{}, err
	}
	user := User{ID: created.ID, Email: created.Email, CreatedAt: created.CreatedAt}

	token, err := auth.SignUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	stored, err := s.repo.GetUserAuthByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrAuthNotFound) {
			return "", User{}, ErrInvalidCredentials
		}
		return "", User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(password)); err != nil {
		return "", User{}, ErrInvalidCredentials
	}
	user := User{
		ID:        stored.User.ID,
		Email:     stored.User.Email,
		CreatedAt: stored.User.CreatedAt,
	}

	token, err := auth.SignUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (User, error) {
	stored, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrAuthNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return User{ID: stored.ID, Email: stored.Email, CreatedAt: stored.CreatedAt}, nil
}
