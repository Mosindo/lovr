package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"example.com/api/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	dbPool    *pgxpool.Pool
	jwtSecret []byte
}

func NewAuthService(dbPool *pgxpool.Pool, jwtSecret []byte) *AuthService {
	return &AuthService{
		dbPool:    dbPool,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", User{}, err
	}

	var user User
	err = s.dbPool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, normalizedEmail, string(hash)).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return "", User{}, ErrEmailExists
		}
		return "", User{}, err
	}

	token, err := auth.SignUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	var user User
	var passwordHash string
	err := s.dbPool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, normalizedEmail).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", User{}, ErrInvalidCredentials
		}
		return "", User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", User{}, ErrInvalidCredentials
	}

	token, err := auth.SignUserToken(s.jwtSecret, user.ID)
	if err != nil {
		return "", User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (User, error) {
	var user User
	err := s.dbPool.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
