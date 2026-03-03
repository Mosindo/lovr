package handlers

import (
	"errors"
	"net/http"
	"time"

	"example.com/api/internal/services"
	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type meResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  meResponse `json:"user"`
}

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logHandlerError(c, "auth.register.bind", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		logHandlerError(c, "auth.register", err)
		switch {
		case errors.Is(err, services.ErrEmailExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		Token: token,
		User: meResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
	logHandlerEvent(c, "auth.register.success", map[string]string{"created_user_id": user.ID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logHandlerError(c, "auth.login.bind", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		logHandlerError(c, "auth.login", err)
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not login"})
		}
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token: token,
		User: meResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
	logHandlerEvent(c, "auth.login.success", map[string]string{"authenticated_user_id": user.ID})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("userID")
	user, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		logHandlerError(c, "auth.me", err)
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch user"})
		}
		return
	}

	c.JSON(http.StatusOK, meResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}
