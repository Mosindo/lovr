package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAuthEmailExists = errors.New("auth email already exists")
	ErrAuthNotFound    = errors.New("auth user not found")
)

type AuthUser struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type AuthUserWithPassword struct {
	User         AuthUser
	PasswordHash string
}

type AuthRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (AuthUser, error)
	GetUserAuthByEmail(ctx context.Context, email string) (AuthUserWithPassword, error)
	GetUserByID(ctx context.Context, userID string) (AuthUser, error)
}

type PGAuthRepository struct {
	dbPool *pgxpool.Pool
}

func NewPGAuthRepository(dbPool *pgxpool.Pool) *PGAuthRepository {
	return &PGAuthRepository{dbPool: dbPool}
}

func (r *PGAuthRepository) CreateUser(ctx context.Context, email, passwordHash string) (AuthUser, error) {
	var user AuthUser
	err := r.dbPool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, email, passwordHash).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return AuthUser{}, ErrAuthEmailExists
		}
		return AuthUser{}, err
	}
	return user, nil
}

func (r *PGAuthRepository) GetUserAuthByEmail(ctx context.Context, email string) (AuthUserWithPassword, error) {
	var user AuthUserWithPassword
	err := r.dbPool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.User.ID, &user.User.Email, &user.PasswordHash, &user.User.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUserWithPassword{}, ErrAuthNotFound
		}
		return AuthUserWithPassword{}, err
	}
	return user, nil
}

func (r *PGAuthRepository) GetUserByID(ctx context.Context, userID string) (AuthUser, error) {
	var user AuthUser
	err := r.dbPool.QueryRow(ctx, `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUser{}, ErrAuthNotFound
		}
		return AuthUser{}, err
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
