package router

import (
	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/internal/handler"
	"github.com/quocbaong/profile-backend/internal/middleware"
)

func SetupRouter(
	cfg *config.Config,
	authH *handler.AuthHandler,
	projectH *handler.ProjectHandler,
	msgH *handler.MessageHandler,
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
		api.POST("/contact", msgH.SubmitContact)
		api.POST("/auth/login", authH.Login)

		// ========================
		// 2. PROTECTED ADMIN ROUTES
		// ========================
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		{
			// Auth & Profile
			admin.GET("/me", authH.GetProfile)
			admin.POST("/change-password", authH.ChangePassword)

			// Projects Management
			admin.GET("/projects", projectH.GetAllProjects)
			admin.GET("/projects/:id", projectH.GetProjectByID)
			admin.POST("/projects", projectH.CreateProject)
			admin.PUT("/projects/:id", projectH.UpdateProject)
			admin.DELETE("/projects/:id", projectH.DeleteProject)

			// Messages Management
			admin.GET("/messages", msgH.GetAllMessages)
			admin.GET("/messages/unread-count", msgH.GetUnreadCount)
			admin.PATCH("/messages/:id/read", msgH.MarkAsRead)
			admin.DELETE("/messages/:id", msgH.DeleteMessage)
		}
	}

	return r
}
