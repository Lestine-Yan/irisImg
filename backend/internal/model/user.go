package model

import "time"

// User 是用户实体，目前用于演示分层架构。
// 后续接入数据库时可以加上 gorm tag。
type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest 是创建用户的入参 DTO。
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=32"`
	Email    string `json:"email"    binding:"required,email"`
}
