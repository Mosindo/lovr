package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"example.com/api/internal/repositories"
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
	repo repositories.ChatRepository
}

const (
	defaultChatMessagesLimit = 200
	maxChatMessagesLimit     = 500
)

func NewChatService(repo repositories.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]ChatSummary, error) {
	chats, err := s.repo.ListChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	payload := make([]ChatSummary, 0, len(chats))
	for _, chat := range chats {
		item := ChatSummary{
			UserID:        chat.UserID,
			UserEmail:     chat.UserEmail,
			UserCreatedAt: chat.UserCreatedAt,
		}
		if chat.LastMessage != nil {
			item.LastMessage = &ChatMessagePreview{
				Content:   chat.LastMessage.Content,
				CreatedAt: chat.LastMessage.CreatedAt,
			}
		}
		payload = append(payload, item)
	}

	return payload, nil
}

func (s *ChatService) ListMessages(ctx context.Context, userID, otherUserID string) ([]ChatMessage, error) {
	return s.ListMessagesWithLimit(ctx, userID, otherUserID, defaultChatMessagesLimit)
}

func (s *ChatService) ListMessagesWithLimit(ctx context.Context, userID, otherUserID string, limit int) ([]ChatMessage, error) {
	if err := s.ensureCanChat(ctx, userID, otherUserID); err != nil {
		return nil, err
	}

	normalizedLimit := normalizeChatMessagesLimit(limit)
	messages, err := s.repo.ListMessages(ctx, userID, otherUserID, normalizedLimit)
	if err != nil {
		return nil, err
	}
	payload := make([]ChatMessage, 0, len(messages))
	for _, message := range messages {
		payload = append(payload, ChatMessage{
			ID:              message.ID,
			SenderUserID:    message.SenderUserID,
			RecipientUserID: message.RecipientUserID,
			Content:         message.Content,
			CreatedAt:       message.CreatedAt,
		})
	}
	return payload, nil
}

func (s *ChatService) SendMessage(ctx context.Context, userID, otherUserID, content string) (ChatMessage, error) {
	if err := s.ensureCanChat(ctx, userID, otherUserID); err != nil {
		return ChatMessage{}, err
	}

	normalizedContent := strings.TrimSpace(content)
	if normalizedContent == "" {
		return ChatMessage{}, ErrMessageContentNeeded
	}

	message, err := s.repo.CreateMessage(ctx, userID, otherUserID, normalizedContent)
	if err != nil {
		return ChatMessage{}, err
	}

	return ChatMessage{
		ID:              message.ID,
		SenderUserID:    message.SenderUserID,
		RecipientUserID: message.RecipientUserID,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt,
	}, nil
}

func (s *ChatService) ensureCanChat(ctx context.Context, userID, otherUserID string) error {
	exists, err := s.repo.UserExists(ctx, otherUserID)
	if err != nil {
		return ErrValidateChatTarget
	}
	if !exists {
		return ErrUserNotFound
	}

	blocked, err := s.repo.UsersBlocked(ctx, userID, otherUserID)
	if err != nil {
		return ErrValidateChatAccess
	}
	if blocked {
		return ErrInteractionBlock
	}

	matched, err := s.repo.UsersMatched(ctx, userID, otherUserID)
	if err != nil {
		return ErrValidateChatAccess
	}
	if !matched {
		return ErrChatRequiresMatch
	}

	return nil
}

func normalizeChatMessagesLimit(limit int) int {
	if limit <= 0 {
		return defaultChatMessagesLimit
	}
	if limit > maxChatMessagesLimit {
		return maxChatMessagesLimit
	}
	return limit
}
