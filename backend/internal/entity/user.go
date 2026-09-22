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
	ID                string    `json:"id" gorm:"primaryKey"`
	GoogleSubject     string    `json:"-" gorm:"column:google_subject;uniqueIndex"`
	Name              string    `json:"name"`
	Email             string    `json:"email" gorm:"uniqueIndex"`
	EmailVerified     bool      `json:"email_verified" gorm:"column:email_verified"`
	AvatarURL         string    `json:"avatar_url" gorm:"column:avatar_url"`
	PasswordHash      string    `json:"-" gorm:"column:password_hash"`
	Timezone          string    `json:"timezone" gorm:"not null;default:Asia/Ho_Chi_Minh"`
	TimezoneConfirmed bool      `json:"timezone_confirmed" gorm:"column:timezone_confirmed;not null;default:false"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
