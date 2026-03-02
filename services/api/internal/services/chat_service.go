package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrChatRequiresMatch    = errors.New("chat allowed only after match")
	ErrValidateChatTarget   = errors.New("could not validate chat target")
	ErrValidateChatAccess   = errors.New("could not validate chat access")
	ErrMessageContentNeeded = errors.New("message content required")
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

type ChatService struct {
	dbPool *pgxpool.Pool
}

func NewChatService(dbPool *pgxpool.Pool) *ChatService {
	return &ChatService{dbPool: dbPool}
}

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]ChatSummary, error) {
	rows, err := s.dbPool.Query(ctx, `
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

func (s *ChatService) ListMessages(ctx context.Context, userID, otherUserID string) ([]ChatMessage, error) {
	if err := s.ensureCanChat(ctx, userID, otherUserID); err != nil {
		return nil, err
	}

	rows, err := s.dbPool.Query(ctx, `
		SELECT id, sender_user_id, recipient_user_id, content, created_at
		FROM messages
		WHERE (sender_user_id = $1 AND recipient_user_id = $2)
		   OR (sender_user_id = $2 AND recipient_user_id = $1)
		ORDER BY created_at ASC
		LIMIT 200
	`, userID, otherUserID)
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

func (s *ChatService) SendMessage(ctx context.Context, userID, otherUserID, content string) (ChatMessage, error) {
	if err := s.ensureCanChat(ctx, userID, otherUserID); err != nil {
		return ChatMessage{}, err
	}

	normalizedContent := strings.TrimSpace(content)
	if normalizedContent == "" {
		return ChatMessage{}, ErrMessageContentNeeded
	}

	var message ChatMessage
	err := s.dbPool.QueryRow(ctx, `
		INSERT INTO messages (sender_user_id, recipient_user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, sender_user_id, recipient_user_id, content, created_at
	`, userID, otherUserID, normalizedContent).Scan(
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

func (s *ChatService) ensureCanChat(ctx context.Context, userID, otherUserID string) error {
	exists, err := s.userExists(ctx, otherUserID)
	if err != nil {
		return ErrValidateChatTarget
	}
	if !exists {
		return ErrUserNotFound
	}

	blocked, err := s.usersBlocked(ctx, userID, otherUserID)
	if err != nil {
		return ErrValidateChatAccess
	}
	if blocked {
		return ErrInteractionBlock
	}

	matched, err := s.usersMatched(ctx, userID, otherUserID)
	if err != nil {
		return ErrValidateChatAccess
	}
	if !matched {
		return ErrChatRequiresMatch
	}

	return nil
}

func (s *ChatService) userExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := s.dbPool.QueryRow(ctx, `
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

func (s *ChatService) usersBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
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

func (s *ChatService) usersMatched(ctx context.Context, userID, otherUserID string) (bool, error) {
	var matched bool
	err := s.dbPool.QueryRow(ctx, `
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
