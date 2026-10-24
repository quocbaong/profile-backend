package database

import (
	"log"

	"github.com/glebarez/sqlite"
	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) *gorm.DB {
	var dialector gorm.Dialector

	if cfg.DBDriver == "postgres" {
		dialector = postgres.Open(cfg.DBDSN)
	} else {
		// Pure-Go SQLite driver (no CGO required)
		dialector = sqlite.Open(cfg.DBDSN)
	}

	dbLogLevel := logger.Warn
	if cfg.AppEnv == "development" {
		dbLogLevel = logger.Info
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(dbLogLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ContactMessage{},
	); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database connection & migration completed successfully")

	// Seed default admin user if not exists
	seedAdmin(db, cfg)

	return db
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.User{}).Where("username = ?", cfg.AdminUsername).Count(&count)
	if count == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash default admin password: %v", err)
			return
		}

		admin := models.User{
			Username: cfg.AdminUsername,
			Password: string(hashed),
			Name:     "Administrator",
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to seed admin user: %v", err)
		} else {
			log.Printf("Default admin user created: %s", cfg.AdminUsername)
		}
	}
}
