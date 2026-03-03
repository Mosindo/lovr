package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"example.com/api/internal/services"
	"github.com/gin-gonic/gin"
)

type chatMessagePreview struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type chatSummaryResponse struct {
	User        meResponse          `json:"user"`
	LastMessage *chatMessagePreview `json:"lastMessage,omitempty"`
}

type chatsResponse struct {
	Chats []chatSummaryResponse `json:"chats"`
}

type sendMessageRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type chatMessageResponse struct {
	ID              string    `json:"id"`
	SenderUserID    string    `json:"senderUserId"`
	RecipientUserID string    `json:"recipientUserId"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
}

type chatMessagesResponse struct {
	Messages []chatMessageResponse `json:"messages"`
}

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) Chats(c *gin.Context) {
	userID := c.GetString("userID")
	chats, err := h.chatService.ListChats(c.Request.Context(), userID)
	if err != nil {
		logHandlerError(c, "chat.list", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch chats"})
		return
	}

	payload := make([]chatSummaryResponse, 0, len(chats))
	for _, chat := range chats {
		item := chatSummaryResponse{
			User: meResponse{
				ID:        chat.UserID,
				Email:     chat.UserEmail,
				CreatedAt: chat.UserCreatedAt,
			},
		}
		if chat.LastMessage != nil {
			item.LastMessage = &chatMessagePreview{
				Content:   chat.LastMessage.Content,
				CreatedAt: chat.LastMessage.CreatedAt,
			}
		}
		payload = append(payload, item)
	}

	c.JSON(http.StatusOK, chatsResponse{Chats: payload})
}

func (h *ChatHandler) ChatMessages(c *gin.Context) {
	userID := c.GetString("userID")
	otherUserID := strings.TrimSpace(c.Param("userId"))
	if !isUUIDLike(otherUserID) {
		logHandlerError(c, "chat.messages.validate_user_id", errors.New("invalid user id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if otherUserID == userID {
		logHandlerError(c, "chat.messages.validate_self", errors.New("cannot open self chat"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open self chat"})
		return
	}

	limit := parsePositiveInt(c.Query("limit"), 200)
	messages, err := h.chatService.ListMessagesWithLimit(c.Request.Context(), userID, otherUserID, limit)
	if err != nil {
		logHandlerError(c, "chat.messages", err)
		switch {
		case errors.Is(err, services.ErrValidateChatTarget):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate chat target"})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, services.ErrValidateChatAccess):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate chat access"})
		case errors.Is(err, services.ErrInteractionBlock):
			c.JSON(http.StatusForbidden, gin.H{"error": "interaction blocked"})
		case errors.Is(err, services.ErrChatRequiresMatch):
			c.JSON(http.StatusForbidden, gin.H{"error": "chat allowed only after match"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch messages"})
		}
		return
	}

	payload := make([]chatMessageResponse, 0, len(messages))
	for _, message := range messages {
		payload = append(payload, chatMessageResponse{
			ID:              message.ID,
			SenderUserID:    message.SenderUserID,
			RecipientUserID: message.RecipientUserID,
			Content:         message.Content,
			CreatedAt:       message.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, chatMessagesResponse{Messages: payload})
}

func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	userID := c.GetString("userID")
	otherUserID := strings.TrimSpace(c.Param("userId"))
	if !isUUIDLike(otherUserID) {
		logHandlerError(c, "chat.send.validate_user_id", errors.New("invalid user id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if otherUserID == userID {
		logHandlerError(c, "chat.send.validate_self", errors.New("cannot message yourself"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot message yourself"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logHandlerError(c, "chat.send.bind", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), userID, otherUserID, req.Content)
	if err != nil {
		logHandlerError(c, "chat.send", err)
		switch {
		case errors.Is(err, services.ErrValidateChatTarget):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate chat target"})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, services.ErrValidateChatAccess):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not validate chat access"})
		case errors.Is(err, services.ErrInteractionBlock):
			c.JSON(http.StatusForbidden, gin.H{"error": "interaction blocked"})
		case errors.Is(err, services.ErrChatRequiresMatch):
			c.JSON(http.StatusForbidden, gin.H{"error": "chat allowed only after match"})
		case errors.Is(err, services.ErrMessageContentNeeded):
			c.JSON(http.StatusBadRequest, gin.H{"error": "message content required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send message"})
		}
		return
	}

	c.JSON(http.StatusCreated, chatMessageResponse{
		ID:              message.ID,
		SenderUserID:    message.SenderUserID,
		RecipientUserID: message.RecipientUserID,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt,
	})
	logHandlerEvent(c, "chat.send.success", map[string]string{
		"recipient_user_id": otherUserID,
		"message_id":        message.ID,
	})
}

func isUUIDLike(v string) bool {
	if len(v) != 36 {
		return false
	}
	hyphenPos := map[int]struct{}{8: {}, 13: {}, 18: {}, 23: {}}
	for i := 0; i < len(v); i++ {
		ch := v[i]
		if _, ok := hyphenPos[i]; ok {
			if ch != '-' {
				return false
			}
			continue
		}
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}
