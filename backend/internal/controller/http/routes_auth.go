package httpapi

import "github.com/gin-gonic/gin"

func registerPublicRoutes(api *gin.RouterGroup, authHandler *AuthHandler) {
	api.POST(routeLoginPath, authHandler.Login)
	authRoutes := api.Group(routeAuthPath)
	authRoutes.GET(routeGooglePath, authHandler.StartGoogleAuth)
	authRoutes.GET(routeGoogleCallbackPath, authHandler.GoogleCallback)
	api.GET(routeHealthPath, healthCheck)
}

func healthCheck(c *gin.Context) { OK(c, gin.H{"status": "ok"}) }
