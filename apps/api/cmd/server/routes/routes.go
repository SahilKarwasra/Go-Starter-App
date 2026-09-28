package routes

import (
	"api/cmd/server/handler"
	"api/cmd/server/middleware"
	"api/cmd/server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	jwtSecret string,
	authHandler *handler.AuthHandler,
) *gin.Engine {
	engine := gin.Default()

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		utils.Success(c, http.StatusOK, "Service is healthy", gin.H{"status": "UP"})
	})

	api := engine.Group("/api/v1")
	{
		// Public Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/sign-up", authHandler.SignUp)
			auth.POST("/sign-in", authHandler.SignIn)
			auth.POST("/send-otp", authHandler.SendOTP)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/refresh-token", authHandler.RefreshToken)
		}

		// Protected routes (Requires Bearer JWT access token)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/auth/me", authHandler.GetProfile)
			protected.POST("/auth/logout", authHandler.Logout)
		}
	}

	return engine
}
