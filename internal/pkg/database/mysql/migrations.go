package database

import (
	"QA-System/internal/models"

	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Survey{},
		&models.Question{},
		&models.Option{},
		&models.Manage{},
		&models.Type{},
	)
}
