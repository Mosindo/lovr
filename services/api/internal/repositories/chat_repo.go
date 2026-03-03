package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatMessagePreview struct {
	Content   string
	CreatedAt time.Time
}

type ChatSummary struct {
	UserID        string
	UserEmail     string
	UserCreatedAt time.Time
	LastMessage   *ChatMessagePreview
}

type ChatMessage struct {
	ID              string
	SenderUserID    string
	RecipientUserID string
	Content         string
	CreatedAt       time.Time
}

type ChatRepository interface {
	ListChats(ctx context.Context, userID string) ([]ChatSummary, error)
	ListMessages(ctx context.Context, userID, otherUserID string, limit int) ([]ChatMessage, error)
	CreateMessage(ctx context.Context, userID, otherUserID, content string) (ChatMessage, error)
	UserExists(ctx context.Context, userID string) (bool, error)
	UsersBlocked(ctx context.Context, userID, otherUserID string) (bool, error)
	UsersMatched(ctx context.Context, userID, otherUserID string) (bool, error)
}

type PGChatRepository struct {
	dbPool *pgxpool.Pool
}

func NewPGChatRepository(dbPool *pgxpool.Pool) *PGChatRepository {
	return &PGChatRepository{dbPool: dbPool}
}

func (r *PGChatRepository) ListChats(ctx context.Context, userID string) ([]ChatSummary, error) {
	rows, err := r.dbPool.Query(ctx, `
		SELECT
			u.id,
			u.email,
			u.created_at,
			lm.content,
			lm.created_at
		FROM likes sent
		JOIN likes received
		  ON received.from_user_id = sent.to_user_id
		 AND received.to_user_id = sent.from_user_id
		JOIN users u
		  ON u.id = sent.to_user_id
		LEFT JOIN LATERAL (
			SELECT m.content, m.created_at
			FROM messages m
			WHERE (m.sender_user_id = $1 AND m.recipient_user_id = sent.to_user_id)
			   OR (m.sender_user_id = sent.to_user_id AND m.recipient_user_id = $1)
			ORDER BY m.created_at DESC
			LIMIT 1
		) lm ON true
		WHERE sent.from_user_id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = sent.to_user_id)
			   OR (b.blocker_user_id = sent.to_user_id AND b.blocked_user_id = $1)
		  )
		ORDER BY COALESCE(lm.created_at, u.created_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]ChatSummary, 0)
	for rows.Next() {
		var chat ChatSummary
		var lastContent sql.NullString
		var lastCreatedAt sql.NullTime
		if err := rows.Scan(&chat.UserID, &chat.UserEmail, &chat.UserCreatedAt, &lastContent, &lastCreatedAt); err != nil {
			return nil, err
		}
		if lastContent.Valid && lastCreatedAt.Valid {
			chat.LastMessage = &ChatMessagePreview{
				Content:   lastContent.String,
				CreatedAt: lastCreatedAt.Time,
			}
		}
		chats = append(chats, chat)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return chats, nil
}

func (r *PGChatRepository) ListMessages(ctx context.Context, userID, otherUserID string, limit int) ([]ChatMessage, error) {
	rows, err := r.dbPool.Query(ctx, `
		SELECT id, sender_user_id, recipient_user_id, content, created_at
		FROM messages
		WHERE (sender_user_id = $1 AND recipient_user_id = $2)
		   OR (sender_user_id = $2 AND recipient_user_id = $1)
		ORDER BY created_at ASC
		LIMIT $3
	`, userID, otherUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var message ChatMessage
		if err := rows.Scan(&message.ID, &message.SenderUserID, &message.RecipientUserID, &message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return messages, nil
}

func (r *PGChatRepository) CreateMessage(ctx context.Context, userID, otherUserID, content string) (ChatMessage, error) {
	var message ChatMessage
	err := r.dbPool.QueryRow(ctx, `
		INSERT INTO messages (sender_user_id, recipient_user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, sender_user_id, recipient_user_id, content, created_at
	`, userID, otherUserID, content).Scan(
		&message.ID,
		&message.SenderUserID,
		&message.RecipientUserID,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		return ChatMessage{}, err
	}
	return message, nil
}

func (r *PGChatRepository) UserExists(ctx context.Context, userID string) (bool, error) {
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

func (r *PGChatRepository) UsersBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
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

func (r *PGChatRepository) UsersMatched(ctx context.Context, userID, otherUserID string) (bool, error) {
	var matched bool
	err := r.dbPool.QueryRow(ctx, `
		SELECT
			EXISTS (
				SELECT 1
				FROM likes
				WHERE from_user_id = $1
				  AND to_user_id = $2
			)
			AND EXISTS (
				SELECT 1
				FROM likes
				WHERE from_user_id = $2
				  AND to_user_id = $1
			)
	`, userID, otherUserID).Scan(&matched)
	if err != nil {
		return false, err
	}
	return matched, nil
}
