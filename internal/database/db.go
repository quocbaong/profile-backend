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

	// Seed default admin user if not exists
	seedAdmin(db, cfg)

	// Seed default data if tables are empty
	seedProjects(db)
	seedProfile(db)
	seedServices(db)
	seedSkills(db)
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

func seedProjects(db *gorm.DB) {
	var count int64
	db.Model(&models.Project{}).Count(&count)
	if count > 0 {
		return
	}

	initialProjects := []models.Project{
		{
			Title:       "ZORA - Distributed E-Commerce Ecosystem",
			Category:    "Web & Mobile Platform",
			Tags:        "Spring Boot • Kafka • React • React Native • PostgreSQL • Docker",
			Description: "Comprehensive distributed e-commerce ecosystem featuring 8+ Spring Boot microservices, Kafka event-driven messaging, multi-gateway payments (PayOS, Stripe, MoMo), WebRTC real-time video/chat, and an AI shopping assistant.",
			Image:       "https://raw.githubusercontent.com/quocbaong/zora-ecommerce-ecosystem/main/ecommerce-frontend/public/architecture.png",
			GithubURL:   "https://github.com/quocbaong/zora-ecommerce-ecosystem",
			LiveURL:     "https://zora-ecommerce-ecosystem.vercel.app",
			SortOrder:   1,
			IsActive:    true,
		},
		{
			Title:       "Train Ticket Sales Management",
			Category:    "Desktop System",
			Tags:        "Java • SQL Server • JDBC",
			Description: "Enterprise ticket booking & seat reservation system with ACID transactional guarantees and revenue reporting.",
			Image:       "/images/project_train_ticket.jpg",
			GithubURL:   "https://github.com/quocbaong/qlbanvetau",
			SortOrder:   2,
			IsActive:    true,
		},
		{
			Title:       "SocialPilot Pro",
			Category:    "Desktop Automation",
			Tags:        "Electron • TypeScript • Node.js • Express",
			Description: "All-in-one social media automation platform for downloading watermark-free videos, batch scanning, and publishing.",
			Image:       "/images/project2.jpg",
			GithubURL:   "https://github.com/quocbaong/SocialPilot",
			SortOrder:   3,
			IsActive:    true,
		},
		{
			Title:       "Food Delivery Website (Food Del)",
			Category:    "Web Platform",
			Tags:        "React.js • Tailwind CSS • REST API",
			Description: "Online food ordering platform with reactive cart management, real-time checkout simulation, and admin menu dashboard.",
			Image:       "/images/project_food_del_clean.jpg",
			GithubURL:   "https://github.com/quocbaong/Food_Delivery_Website",
			LiveURL:     "https://fooddeliverywebsite-two.vercel.app/",
			SortOrder:   4,
			IsActive:    true,
		},
		{
			Title:       "Traveling Tour Explorer App",
			Category:    "Mobile App",
			Tags:        "React Native • Expo • MongoDB • REST API",
			Description: "Mobile travel guide helping users search for tourist destinations, plan customized itineraries, and book tours.",
			Image:       "/images/project_traveling.jpg",
			GithubURL:   "https://github.com/quocbaong/Travelling-App",
			SortOrder:   5,
			IsActive:    true,
		},
	}

	for _, p := range initialProjects {
		if err := db.Create(&p).Error; err != nil {
			log.Printf("Failed to seed project %s: %v", p.Title, err)
		}
	}
	log.Printf("Successfully auto-seeded %d projects into database", len(initialProjects))
}

func seedProfile(db *gorm.DB) {
	var count int64
	db.Model(&models.ProfileSettings{}).Count(&count)
	if count > 0 {
		return
	}

	profile := models.ProfileSettings{
		Name:             "Nguyễn Quốc Bảo",
		Title:            "Full Stack Developer",
		Subtitle:         "I am Nguyễn Quốc Bảo",
		AboutTitle:       "Professional And Dedicated Full Stack Developer",
		AboutDescription: "I’m a fourth-year IT student at Industrial University of Ho Chi Minh City (IUH) on my journey to becoming a professional Software Engineer. I’m passionate about modern software architectures, enterprise backend systems, and crafting responsive web and mobile digital experiences that solve real-world problems.\n\nThrough hands-on projects spanning full-featured web platforms, desktop database systems, and cross-platform mobile apps, I have developed strong logical thinking, rapid problem-solving capabilities, and effective team collaboration skills.",
		Email:            "nbao76296@gmail.com",
		Phone:            "+84 399 609 008",
		Address:          "Ho Chi Minh City, Vietnam",
		HeroImage:        "/images/user_slide.jpg",
		AboutImage:       "/images/profile.jpg",
		ResumeURL:        "#",
		GithubURL:        "https://github.com/quocbaong",
		LinkedinURL:      "https://www.linkedin.com/in/ngquocbao-dev106/",
		FacebookURL:      "https://www.facebook.com/qbao.henry",
		InstagramURL:     "https://www.instagram.com/quocbao.106/",
	}

	if err := db.Create(&profile).Error; err != nil {
		log.Printf("Failed to seed profile settings: %v", err)
	} else {
		log.Println("Successfully auto-seeded default profile settings")
	}
}

func seedServices(db *gorm.DB) {
	var count int64
	db.Model(&models.ServiceItem{}).Count(&count)
	if count > 0 {
		return
	}

	services := []models.ServiceItem{
		{
			Title:       "Web Development",
			IconClass:   "fi flaticon-laptop",
			Description: "Crafting responsive, high-performance web applications using React, Next.js, TypeScript, and modern frontend architectures.",
			SortOrder:   1,
			IsActive:    true,
		},
		{
			Title:       "Backend & APIs",
			IconClass:   "fi flaticon-web-design",
			Description: "Engineering resilient RESTful APIs, scalable microservices, and secure authentication pipelines with Java Spring Boot and Node.js.",
			SortOrder:   2,
			IsActive:    true,
		},
		{
			Title:       "Mobile App Development",
			IconClass:   "fi flaticon-smartphone",
			Description: "Developing seamless cross-platform mobile experiences for iOS and Android powered by React Native and Expo.",
			SortOrder:   3,
			IsActive:    true,
		},
		{
			Title:       "Database & Architecture",
			IconClass:   "fi flaticon-verified",
			Description: "Designing ACID-compliant relational schemas with Microsoft SQL Server and flexible document storage with MongoDB.",
			SortOrder:   4,
			IsActive:    true,
		},
	}

	for _, s := range services {
		_ = db.Create(&s).Error
	}
	log.Printf("Successfully auto-seeded %d services", len(services))
}

func seedSkills(db *gorm.DB) {
	var count int64
	db.Model(&models.SkillItem{}).Count(&count)
	if count > 0 {
		return
	}

	skills := []models.SkillItem{
		{Name: "Java", Category: "Backend", Color: "#F89820", SortOrder: 1, IsActive: true},
		{Name: "Spring Boot", Category: "Backend", Color: "#6DB33F", SortOrder: 2, IsActive: true},
		{Name: "React.js", Category: "Frontend", Color: "#61DAFB", SortOrder: 3, IsActive: true},
		{Name: "JavaScript", Category: "Language", Color: "#F7DF1E", SortOrder: 4, IsActive: true},
		{Name: "Express.js", Category: "Backend", Color: "#E8E8E8", SortOrder: 5, IsActive: true},
		{Name: "SQL Server", Category: "Database", Color: "#CC292B", SortOrder: 6, IsActive: true},
		{Name: "MongoDB", Category: "Database", Color: "#47A248", SortOrder: 7, IsActive: true},
		{Name: "Docker", Category: "DevOps", Color: "#2496ED", SortOrder: 8, IsActive: true},
		{Name: "Git & GitHub", Category: "Tools", Color: "#ffffff", SortOrder: 9, IsActive: true},
	}

	for _, sk := range skills {
		_ = db.Create(&sk).Error
	}
	log.Printf("Successfully auto-seeded %d skills", len(skills))
}
