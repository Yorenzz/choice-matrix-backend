package routes

import (
	"choice-matrix-backend/internal/api/handlers"
	"choice-matrix-backend/internal/api/middleware"
	"choice-matrix-backend/internal/auth"
	"choice-matrix-backend/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, accessTokenTTL, refreshTokenTTL time.Duration) {
	userRepo := repository.NewUserRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	matrixRepo := repository.NewMatrixRepository(db)
	refreshStore := auth.NewRefreshStore(redisClient)

	authHandler := handlers.NewAuthHandler(userRepo, refreshStore, accessTokenTTL, refreshTokenTTL)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceRepo)
	matrixHandler := handlers.NewMatrixHandler(matrixRepo, workspaceRepo)
	aiHandler := handlers.NewAIHandler(workspaceRepo, matrixRepo)

	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":     "ok",
				"database_up": true,
				"database_mode": gin.H{
					"connected": true,
				},
			})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			protected.GET("/folders", workspaceHandler.GetFolders)
			protected.POST("/folders", workspaceHandler.CreateFolder)
			protected.GET("/projects", workspaceHandler.GetProjects)
			protected.POST("/projects", workspaceHandler.CreateProject)
			protected.GET("/projects/:id", workspaceHandler.GetProjectDetails)
			protected.PUT("/projects/:id", workspaceHandler.UpdateProject)
			protected.DELETE("/projects/:id", workspaceHandler.DeleteProject)

			protected.GET("/projects/:id/payload", matrixHandler.GetProjectPayload)
			protected.POST("/projects/:id/rows", matrixHandler.CreateRow)
			protected.PUT("/projects/:id/rows/:rowId", matrixHandler.UpdateRow)
			protected.DELETE("/projects/:id/rows/:rowId", matrixHandler.DeleteRow)
			protected.PUT("/projects/:id/rows/reorder", matrixHandler.ReorderRows)
			protected.POST("/projects/:id/columns", matrixHandler.CreateColumn)
			protected.PUT("/projects/:id/columns/:columnId", matrixHandler.UpdateColumn)
			protected.DELETE("/projects/:id/columns/:columnId", matrixHandler.DeleteColumn)
			protected.PUT("/projects/:id/columns/reorder", matrixHandler.ReorderColumns)
			protected.PUT("/projects/:id/cells", matrixHandler.UpsertCell)

			protected.POST("/projects/:id/ai/summary", aiHandler.GenerateSummary)
		}
	}
}
