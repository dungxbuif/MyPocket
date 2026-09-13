package httpapi

import "github.com/gin-gonic/gin"

func registerAccountRoutes(protected *gin.RouterGroup, profileHandler *ProfileHandler, homeHandler *HomeHandler) {
	protected.GET(routeAuthPath+routeProfilePath, profileHandler.GetProfile)
	protected.GET(routeHomePath, homeHandler.GetHome)
}
