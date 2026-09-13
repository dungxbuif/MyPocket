package db

import (
	"errors"
	"fmt"

	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
)

func EnsureSchema(db *gorm.DB) error {
	if db.Migrator().HasTable("app_users") && !db.Migrator().HasTable("user") {
		if err := db.Migrator().RenameTable("app_users", "user"); err != nil {
			return fmt.Errorf("rename app_users to user: %w", err)
		}
	}
	if err := db.AutoMigrate(&entity.User{}, &entity.Wallet{}, &entity.Category{}, &entity.CategoryWallet{}); err != nil {
		return err
	}
	return SeedSystemCategories(db)
}

func SeedSystemCategories(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, seed := range entity.SystemCategorySeeds() {
			category := entity.Category{ID: seed.ID, Kind: seed.Kind, Name: seed.Name, IsSystem: true}
			category.SystemKey = &seed.SystemKey
			if seed.ParentKey != "" {
				var parent entity.Category
				if err := tx.Where("system_key = ?", seed.ParentKey).First(&parent).Error; err != nil {
					return err
				}
				category.ParentID = &parent.ID
			}
			var existing entity.Category
			if err := tx.Where("system_key = ?", seed.SystemKey).First(&existing).Error; err == nil {
				existing.Name, existing.Kind, existing.IsSystem, existing.ParentID = category.Name, category.Kind, true, category.ParentID
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&category).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
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
