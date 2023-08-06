package database

import (
	"QA-system/model"
	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		model.List{},
		model.Question{},
		model.Ans{},
	)
}
