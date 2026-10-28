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

	// Auto-backfill skill categories & sub-groups if needed
	backfillSkillGroups(db)

	return db
}

func backfillSkillGroups(db *gorm.DB) {
	// 1. Cloud/DevOps -> Database: SQL Server, MongoDB, PostgreSQL, Redis, etc.
	db.Model(&models.SkillItem{}).
		Where("LOWER(name) LIKE '%sql%' OR LOWER(name) LIKE '%mongo%' OR LOWER(name) LIKE '%redis%' OR LOWER(name) LIKE '%postgres%' OR LOWER(name) LIKE '%minio%' OR LOWER(category) = 'database'").
		Updates(map[string]interface{}{
			"section_group": "cloud_devops",
			"sub_group":     "database",
		})

	// 2. Cloud/DevOps -> Cloud, DevOps & Tools: Docker, Git, Firebase, etc.
	db.Model(&models.SkillItem{}).
		Where("LOWER(name) LIKE '%docker%' OR LOWER(name) LIKE '%git%' OR LOWER(name) LIKE '%firebase%' OR LOWER(name) LIKE '%supabase%' OR LOWER(name) LIKE '%linux%' OR LOWER(name) LIKE '%ci%' OR LOWER(category) = 'devops' OR LOWER(category) = 'tools'").
		Updates(map[string]interface{}{
			"section_group": "cloud_devops",
			"sub_group":     "cloud_devops",
		})

	// 3. Frameworks/Libraries -> Frontend: React, Next, Vue, etc.
	db.Model(&models.SkillItem{}).
		Where("LOWER(name) LIKE '%react%' OR LOWER(name) LIKE '%next%' OR LOWER(name) LIKE '%vue%' OR LOWER(name) LIKE '%flutter%' OR LOWER(name) LIKE '%tailwind%' OR LOWER(name) LIKE '%redux%' OR LOWER(name) LIKE '%ant design%' OR (LOWER(category) = 'frontend' AND LOWER(name) NOT LIKE '%script%')").
		Updates(map[string]interface{}{
			"section_group": "frameworks_libraries",
			"sub_group":     "frontend",
		})

	// 4. Frameworks/Libraries -> Backend: Spring Boot, Express, NestJS, etc.
	db.Model(&models.SkillItem{}).
		Where("(LOWER(name) LIKE '%spring%' OR LOWER(name) LIKE '%express%' OR LOWER(name) LIKE '%nest%' OR LOWER(name) LIKE '%node%' OR LOWER(name) LIKE '%fastapi%' OR LOWER(name) LIKE '%django%' OR LOWER(category) = 'backend') AND LOWER(name) NOT LIKE '%java%' AND LOWER(name) NOT LIKE '%go%'").
		Updates(map[string]interface{}{
			"section_group": "frameworks_libraries",
			"sub_group":     "backend",
		})

	// 5. Languages: Java, JavaScript, TypeScript, Go, etc. (Always takes precedence)
	db.Model(&models.SkillItem{}).
		Where("(LOWER(name) LIKE '%java%' AND LOWER(name) NOT LIKE '%script%') OR LOWER(name) LIKE '%javascript%' OR LOWER(name) LIKE '%typescript%' OR LOWER(name) = 'go' OR LOWER(name) LIKE '%golang%' OR LOWER(category) = 'language'").
		Updates(map[string]interface{}{
			"section_group": "languages",
			"sub_group":     "languages",
		})
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
