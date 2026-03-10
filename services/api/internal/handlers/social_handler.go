package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/api/internal/services"
	"github.com/gin-gonic/gin"
)

type discoverUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type discoverResponse struct {
	Users []discoverUserResponse `json:"users"`
}

type likeRequest struct {
	ToUserID string `json:"toUserId" binding:"required,uuid"`
}

type likeResponse struct {
	Matched bool `json:"matched"`
}

type blockRequest struct {
	ToUserID string `json:"toUserId" binding:"required,uuid"`
}

type blockResponse struct {
	Blocked bool `json:"blocked"`
}

type matchesResponse struct {
	Matches []discoverUserResponse `json:"matches"`
}

type SocialHandler struct {
	socialService *services.SocialService
}

func NewSocialHandler(socialService *services.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

func (h *SocialHandler) Discover(c *gin.Context) {
	userID := c.GetString("userID")
	limit := parsePositiveInt(c.Query("limit"), 50)
	users, err := h.socialService.DiscoverWithLimit(c.Request.Context(), userID, limit)
	if err != nil {
		logHandlerError(c, "social.discover", http.StatusInternalServerError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch discover users"})
		return
	}

	payload := make([]discoverUserResponse, 0, len(users))
	for _, user := range users {
		payload = append(payload, discoverUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, discoverResponse{Users: payload})
}

func (h *SocialHandler) Like(c *gin.Context) {
	userID := c.GetString("userID")

	var req likeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logHandlerError(c, "social.like.bind", http.StatusBadRequest, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	matched, err := h.socialService.Like(c.Request.Context(), userID, strings.TrimSpace(req.ToUserID))
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrCannotLikeSelf):
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "cannot like yourself"})
		case errors.Is(err, services.ErrUserNotFound):
			status = http.StatusNotFound
			c.JSON(status, gin.H{"error": "user not found"})
		case errors.Is(err, services.ErrInteractionBlock):
			status = http.StatusForbidden
			c.JSON(status, gin.H{"error": "interaction blocked"})
		default:
			c.JSON(status, gin.H{"error": "could not save like"})
		}
		logHandlerError(c, "social.like", status, err)
		return
	}

	status := http.StatusOK
	c.JSON(status, likeResponse{Matched: matched})
	logHandlerEvent(c, "social.like.success", status, map[string]string{
		"to_user_id": req.ToUserID,
		"matched":    strconv.FormatBool(matched),
	})
}

func (h *SocialHandler) Block(c *gin.Context) {
	userID := c.GetString("userID")

	var req blockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logHandlerError(c, "social.block.bind", http.StatusBadRequest, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.socialService.Block(c.Request.Context(), userID, strings.TrimSpace(req.ToUserID))
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrCannotBlockSelf):
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "cannot block yourself"})
		case errors.Is(err, services.ErrUserNotFound):
			status = http.StatusNotFound
			c.JSON(status, gin.H{"error": "user not found"})
		default:
			c.JSON(status, gin.H{"error": "could not block user"})
		}
		logHandlerError(c, "social.block", status, err)
		return
	}

	status := http.StatusOK
	c.JSON(status, blockResponse{Blocked: true})
	logHandlerEvent(c, "social.block.success", status, map[string]string{
		"blocked_user_id": req.ToUserID,
	})
}

func (h *SocialHandler) Matches(c *gin.Context) {
	userID := c.GetString("userID")
	matches, err := h.socialService.Matches(c.Request.Context(), userID)
	if err != nil {
		logHandlerError(c, "social.matches", http.StatusInternalServerError, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch matches"})
		return
	}

	payload := make([]discoverUserResponse, 0, len(matches))
	for _, user := range matches {
		payload = append(payload, discoverUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, matchesResponse{Matches: payload})
}

func parsePositiveInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
