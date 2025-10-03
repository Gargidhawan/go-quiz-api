package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"quizapi/internal/models"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect MySQL: %v", err)
	}
	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
	if err := db.AutoMigrate(
		&models.Quiz{},
		&models.Question{},
		&models.Option{},
		&models.Submission{},
		&models.Answer{},
		&models.AnswerOption{},
		&models.User{},
	); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}
	return db
}
