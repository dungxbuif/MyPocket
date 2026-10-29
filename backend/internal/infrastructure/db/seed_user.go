package db

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
)

// SeedDefaultUser is development convenience data. Database structure and
// system categories are exclusively created by versioned SQL migrations.
func SeedDefaultUser(db *gorm.DB, user *entity.User) error {
	var existing entity.User
	if err := db.Where("email = ?", user.Email).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return db.Create(user).Error
	}
	return nil
}
