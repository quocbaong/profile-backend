package main

import (
	"log"

	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/internal/database"
	"github.com/quocbaong/profile-backend/internal/handler"
	"github.com/quocbaong/profile-backend/internal/repository"
	"github.com/quocbaong/profile-backend/internal/router"
	"github.com/quocbaong/profile-backend/internal/service"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database & run AutoMigrate
	db := database.InitDB(cfg)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	skillRepo := repository.NewSkillRepository(db)

	// 4. Initialize Services
	emailService := service.NewEmailService(cfg)
	authService := service.NewAuthService(userRepo, cfg)
	projectService := service.NewProjectService(projectRepo)
	msgService := service.NewMessageService(msgRepo, emailService)
	profileService := service.NewProfileService(profileRepo)
	serviceService := service.NewServiceService(serviceRepo)
	skillService := service.NewSkillService(skillRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	msgHandler := handler.NewMessageHandler(msgService)
	profileHandler := handler.NewProfileHandler(profileService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	skillHandler := handler.NewSkillHandler(skillService)

	// 6. Setup Router
	r := router.SetupRouter(
		cfg,
		authHandler,
		projectHandler,
		msgHandler,
		profileHandler,
		serviceHandler,
		skillHandler,
		projectRepo,
		msgRepo,
	)

	// 7. Start HTTP Server
	log.Printf("Starting server on port :%s (Mode: %s)...", cfg.Port, cfg.AppEnv)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
