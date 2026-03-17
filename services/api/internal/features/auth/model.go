package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type MeResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  MeResponse `json:"user"`
}

type UserClaims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}
