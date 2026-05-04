package database

import (
	"wardrobe/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	err = db.AutoMigrate(&models.Cloth{}, &models.Setting{}, &models.UploadTask{})
	if err != nil {
		return nil, err
	}
	
	return db, nil
}
