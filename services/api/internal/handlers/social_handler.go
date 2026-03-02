package handlers

import (
	"errors"
	"net/http"
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
	users, err := h.socialService.Discover(c.Request.Context(), userID)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	matched, err := h.socialService.Like(c.Request.Context(), userID, strings.TrimSpace(req.ToUserID))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCannotLikeSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot like yourself"})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, services.ErrInteractionBlock):
			c.JSON(http.StatusForbidden, gin.H{"error": "interaction blocked"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save like"})
		}
		return
	}

	c.JSON(http.StatusOK, likeResponse{Matched: matched})
}

func (h *SocialHandler) Block(c *gin.Context) {
	userID := c.GetString("userID")

	var req blockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.socialService.Block(c.Request.Context(), userID, strings.TrimSpace(req.ToUserID))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCannotBlockSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		}
		return
	}

	c.JSON(http.StatusOK, blockResponse{Blocked: true})
}

func (h *SocialHandler) Matches(c *gin.Context) {
	userID := c.GetString("userID")
	matches, err := h.socialService.Matches(c.Request.Context(), userID)
	if err != nil {
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
