package routes

import (
	"choice-matrix-backend/internal/api/handlers"
	"choice-matrix-backend/internal/api/middleware"
	"choice-matrix-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	matrixRepo := repository.NewMatrixRepository(db)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceRepo)
	matrixHandler := handlers.NewMatrixHandler(matrixRepo, workspaceRepo)
	aiHandler := handlers.NewAIHandler()

	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// Public Routes (Auth)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			// Workspace
			protected.GET("/folders", workspaceHandler.GetFolders)
			protected.POST("/folders", workspaceHandler.CreateFolder)
			protected.GET("/projects", workspaceHandler.GetProjects)
			protected.POST("/projects", workspaceHandler.CreateProject)

			// Matrix Canvas
			protected.GET("/projects/:id/payload", matrixHandler.GetProjectPayload)

			protected.POST("/projects/:id/rows", matrixHandler.CreateRow)
			protected.PUT("/projects/:id/rows/reorder", matrixHandler.ReorderRows)

			protected.POST("/projects/:id/columns", matrixHandler.CreateColumn)
			protected.PUT("/projects/:id/columns/reorder", matrixHandler.ReorderColumns)

			protected.PUT("/projects/:id/cells", matrixHandler.UpsertCell)

			// AI
			protected.POST("/projects/:id/ai/summary", aiHandler.GenerateSummary)
		}
	}
}
