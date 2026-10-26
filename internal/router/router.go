package router

import (
	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/internal/handler"
	"github.com/quocbaong/profile-backend/internal/middleware"
	"github.com/quocbaong/profile-backend/internal/repository"
)

func SetupRouter(
	cfg *config.Config,
	authH *handler.AuthHandler,
	projectH *handler.ProjectHandler,
	msgH *handler.MessageHandler,
	profileH *handler.ProfileHandler,
	serviceH *handler.ServiceHandler,
	skillH *handler.SkillHandler,
	projectRepo repository.ProjectRepository,
	msgRepo repository.MessageRepository,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware(cfg))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "quocbao-portfolio-backend",
		})
	})

	api := r.Group("/api")
	{
		// ========================
		// 1. PUBLIC ROUTES
		// ========================
		api.GET("/projects", projectH.GetPublicProjects)
		api.POST("/contact", middleware.ContactRateLimitMiddleware(), msgH.SubmitContact)
		api.POST("/auth/login", authH.Login)
		api.GET("/profile", profileH.GetProfile)
		api.GET("/services", serviceH.GetPublicServices)
		api.GET("/skills", skillH.GetPublicSkills)

		// ========================
		// 2. PROTECTED ADMIN ROUTES
		// ========================
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		{
			// Auth & Account
			admin.GET("/me", authH.GetProfile)
			admin.POST("/change-password", authH.ChangePassword)

			// Dashboard Stats & Uploads
			dashboardH := handler.NewDashboardHandler(projectRepo, msgRepo)
			uploadH := handler.NewUploadHandler(cfg)
			admin.GET("/dashboard/stats", dashboardH.GetStats)
			admin.POST("/upload", uploadH.UploadFile)

			// Portfolio Profile Information Settings
			admin.GET("/profile", profileH.GetProfile)
			admin.PUT("/profile", profileH.UpdateProfile)

			// Projects Management
			admin.GET("/projects", projectH.GetAllProjects)
			admin.GET("/projects/:id", projectH.GetProjectByID)
			admin.POST("/projects", projectH.CreateProject)
			admin.PUT("/projects/:id", projectH.UpdateProject)
			admin.DELETE("/projects/:id", projectH.DeleteProject)

			// Services Management
			admin.GET("/services", serviceH.GetAllServices)
			admin.GET("/services/:id", serviceH.GetServiceByID)
			admin.POST("/services", serviceH.CreateService)
			admin.PUT("/services/:id", serviceH.UpdateService)
			admin.DELETE("/services/:id", serviceH.DeleteService)

			// Skills Management
			admin.GET("/skills", skillH.GetAllSkills)
			admin.GET("/skills/:id", skillH.GetSkillByID)
			admin.POST("/skills", skillH.CreateSkill)
			admin.PUT("/skills/:id", skillH.UpdateSkill)
			admin.DELETE("/skills/:id", skillH.DeleteSkill)

			// Messages Management
			admin.GET("/messages", msgH.GetAllMessages)
			admin.GET("/messages/unread-count", msgH.GetUnreadCount)
			admin.PATCH("/messages/:id/read", msgH.MarkAsRead)
			admin.DELETE("/messages/:id", msgH.DeleteMessage)
		}
	}

	return r
}
