package httpapi

import "github.com/gin-gonic/gin"

func registerAccountRoutes(protected *gin.RouterGroup, profileHandler *ProfileHandler, homeHandler *HomeHandler) {
	protected.GET(routeAuthPath+routeProfilePath, profileHandler.GetProfile)
	protected.PATCH(routeAuthPath+routeProfilePath, profileHandler.UpdateTimezone)
	protected.GET(routeHomePath, homeHandler.GetHome)
}
