package models

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	gorm.Model
	Title   string `gorm:"uniqueIndex;not null" json:"title,omitempty"`
	Content string `gorm:"not null" json:"content,omitempty"`
}

type CreatePostInput struct {
	Title     string    `json:"title"  binding:"required"`
	Content   string    `json:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type UpdatePostInput struct {
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	CreateAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
