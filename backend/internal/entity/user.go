package entity

import "time"

type GoogleProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
}

type User struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	GoogleSubject string    `json:"-" gorm:"column:google_subject;uniqueIndex"`
	Name          string    `json:"name"`
	Email         string    `json:"email" gorm:"uniqueIndex"`
	EmailVerified bool      `json:"email_verified" gorm:"column:email_verified"`
	AvatarURL     string    `json:"avatar_url" gorm:"column:avatar_url"`
	PasswordHash  string    `json:"-" gorm:"column:password_hash"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
