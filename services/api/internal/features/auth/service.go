package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret []byte) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", User{}, err
	}

	created, err := s.repo.CreateUser(ctx, normalizedEmail, string(hash))
	if err != nil {
		if errors.Is(err, ErrRepositoryEmailExists) {
			return "", User{}, ErrEmailExists
		}
		return "", User{}, err
	}
	user := User{ID: created.ID, Email: created.Email, CreatedAt: created.CreatedAt}

	token, err := signUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	stored, err := s.repo.GetUserAuthByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, ErrRepositoryNotFound) {
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

	token, err := signUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *Service) Me(ctx context.Context, userID string) (User, error) {
	stored, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrRepositoryNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return User{ID: stored.ID, Email: stored.Email, CreatedAt: stored.CreatedAt}, nil
}

func signUserToken(secret []byte, userID string) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
