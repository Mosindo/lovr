package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SocialUser struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type SocialRepository interface {
	ListDiscoverUsers(ctx context.Context, userID string, limit int) ([]SocialUser, error)
	UserExists(ctx context.Context, userID string) (bool, error)
	UsersBlocked(ctx context.Context, userID, otherUserID string) (bool, error)
	InsertLike(ctx context.Context, fromUserID, toUserID string) error
	HasReciprocalLike(ctx context.Context, fromUserID, toUserID string) (bool, error)
	BlockAndDeleteLikes(ctx context.Context, blockerUserID, blockedUserID string) error
	ListMatches(ctx context.Context, userID string) ([]SocialUser, error)
}

type PGSocialRepository struct {
	dbPool *pgxpool.Pool
}

func NewPGSocialRepository(dbPool *pgxpool.Pool) *PGSocialRepository {
	return &PGSocialRepository{dbPool: dbPool}
}

func (r *PGSocialRepository) ListDiscoverUsers(ctx context.Context, userID string, limit int) ([]SocialUser, error) {
	rows, err := r.dbPool.Query(ctx, `
		SELECT u.id, u.email, u.created_at
		FROM users u
		WHERE u.id <> $1
		  AND NOT EXISTS (
			SELECT 1
			FROM likes l
			WHERE l.from_user_id = $1
			  AND l.to_user_id = u.id
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = u.id)
			   OR (b.blocker_user_id = u.id AND b.blocked_user_id = $1)
		  )
		ORDER BY u.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]SocialUser, 0)
	for rows.Next() {
		var user SocialUser
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (r *PGSocialRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PGSocialRepository) UsersBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
	var blocked bool
	err := r.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM blocks
			WHERE (blocker_user_id = $1 AND blocked_user_id = $2)
			   OR (blocker_user_id = $2 AND blocked_user_id = $1)
		)
	`, userID, otherUserID).Scan(&blocked)
	if err != nil {
		return false, err
	}
	return blocked, nil
}

func (r *PGSocialRepository) InsertLike(ctx context.Context, fromUserID, toUserID string) error {
	_, err := r.dbPool.Exec(ctx, `
		INSERT INTO likes (from_user_id, to_user_id)
		VALUES ($1, $2)
		ON CONFLICT (from_user_id, to_user_id) DO NOTHING
	`, fromUserID, toUserID)
	return err
}

func (r *PGSocialRepository) HasReciprocalLike(ctx context.Context, fromUserID, toUserID string) (bool, error) {
	var matched bool
	err := r.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM likes
			WHERE from_user_id = $1
			  AND to_user_id = $2
		)
	`, toUserID, fromUserID).Scan(&matched)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func (r *PGSocialRepository) BlockAndDeleteLikes(ctx context.Context, blockerUserID, blockedUserID string) error {
	tx, err := r.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO blocks (blocker_user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_user_id, blocked_user_id) DO NOTHING
	`, blockerUserID, blockedUserID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM likes
		WHERE (from_user_id = $1 AND to_user_id = $2)
		   OR (from_user_id = $2 AND to_user_id = $1)
	`, blockerUserID, blockedUserID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PGSocialRepository) ListMatches(ctx context.Context, userID string) ([]SocialUser, error) {
	rows, err := r.dbPool.Query(ctx, `
		SELECT u.id, u.email, u.created_at
		FROM likes sent
		JOIN likes received
		  ON received.from_user_id = sent.to_user_id
		 AND received.to_user_id = sent.from_user_id
		JOIN users u
		  ON u.id = sent.to_user_id
		WHERE sent.from_user_id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = sent.to_user_id)
			   OR (b.blocker_user_id = sent.to_user_id AND b.blocked_user_id = $1)
		  )
		ORDER BY sent.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]SocialUser, 0)
	for rows.Next() {
		var user SocialUser
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		matches = append(matches, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return matches, nil
}
