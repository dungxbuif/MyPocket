package db

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
)

func EnsureSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		return err
	}
	return nil
}

func SeedDefaultUser(db *gorm.DB, user *entity.User) error {
	var existing entity.User
	if err := db.Where("email = ?", user.Email).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.Create(user).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}
