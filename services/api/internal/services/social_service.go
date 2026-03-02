package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCannotLikeSelf   = errors.New("cannot like yourself")
	ErrCannotBlockSelf  = errors.New("cannot block yourself")
	ErrInteractionBlock = errors.New("interaction blocked")
)

type DiscoverUser struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type SocialService struct {
	dbPool *pgxpool.Pool
}

func NewSocialService(dbPool *pgxpool.Pool) *SocialService {
	return &SocialService{dbPool: dbPool}
}

func (s *SocialService) Discover(ctx context.Context, userID string) ([]DiscoverUser, error) {
	rows, err := s.dbPool.Query(ctx, `
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
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]DiscoverUser, 0)
	for rows.Next() {
		var user DiscoverUser
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

func (s *SocialService) Like(ctx context.Context, fromUserID, toUserID string) (bool, error) {
	normalizedToUserID := strings.TrimSpace(toUserID)
	if normalizedToUserID == fromUserID {
		return false, ErrCannotLikeSelf
	}

	exists, err := s.userExists(ctx, normalizedToUserID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrUserNotFound
	}

	blocked, err := s.usersBlocked(ctx, fromUserID, normalizedToUserID)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, ErrInteractionBlock
	}

	if _, err := s.dbPool.Exec(ctx, `
		INSERT INTO likes (from_user_id, to_user_id)
		VALUES ($1, $2)
		ON CONFLICT (from_user_id, to_user_id) DO NOTHING
	`, fromUserID, normalizedToUserID); err != nil {
		return false, err
	}

	var matched bool
	err = s.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM likes
			WHERE from_user_id = $1
			  AND to_user_id = $2
		)
	`, normalizedToUserID, fromUserID).Scan(&matched)
	if err != nil {
		return false, err
	}

	return matched, nil
}

func (s *SocialService) Block(ctx context.Context, blockerUserID, blockedUserID string) error {
	normalizedBlockedUserID := strings.TrimSpace(blockedUserID)
	if normalizedBlockedUserID == blockerUserID {
		return ErrCannotBlockSelf
	}

	exists, err := s.userExists(ctx, normalizedBlockedUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO blocks (blocker_user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_user_id, blocked_user_id) DO NOTHING
	`, blockerUserID, normalizedBlockedUserID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM likes
		WHERE (from_user_id = $1 AND to_user_id = $2)
		   OR (from_user_id = $2 AND to_user_id = $1)
	`, blockerUserID, normalizedBlockedUserID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *SocialService) Matches(ctx context.Context, userID string) ([]DiscoverUser, error) {
	rows, err := s.dbPool.Query(ctx, `
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

	matches := make([]DiscoverUser, 0)
	for rows.Next() {
		var user DiscoverUser
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

func (s *SocialService) userExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := s.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

func (s *SocialService) usersBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
	var blocked bool
	err := s.dbPool.QueryRow(ctx, `
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
