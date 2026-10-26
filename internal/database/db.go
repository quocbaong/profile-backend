package database

import (
	"log"
	"time"

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
		// Compatible with Supabase direct connection (port 5432) and PgBouncer pooler (port 6543)
		dialector = postgres.New(postgres.Config{
			DSN:                  cfg.DBDSN,
			PreferSimpleProtocol: true, // Disables prepared statement caching for PgBouncer / Supabase transaction pooler
		})
		log.Println("Connecting to PostgreSQL database (Supabase)...")
	} else {
		// Pure-Go SQLite driver (no CGO required)
		dialector = sqlite.Open(cfg.DBDSN)
		log.Printf("Connecting to SQLite database (%s)...", cfg.DBDSN)
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

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ContactMessage{},
		&models.ProfileSettings{},
		&models.ServiceItem{},
		&models.SkillItem{},
	); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database connection & migration completed successfully")

	// Seed admin user if configured in .env and not existing in DB
	seedAdmin(db, cfg)

	return db
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return
	}

	var count int64
	db.Model(&models.User{}).Where("username = ?", cfg.AdminUsername).Count(&count)
	if count == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash admin password: %v", err)
			return
		}

		admin := models.User{
			Username: cfg.AdminUsername,
			Password: string(hashed),
			Name:     "Administrator",
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Printf("Admin user ready: %s", cfg.AdminUsername)
		}
	}
}
