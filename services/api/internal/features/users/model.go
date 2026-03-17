package users

import "time"

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}
