package httpapi

import "github.com/gin-gonic/gin"

func registerCategoryRoutes(protected *gin.RouterGroup, handler *CategoryHandler) {
	routes := protected.Group(routeCategoriesPath)
	routes.GET(routeCollectionPath, handler.ListCategories)
	routes.POST(routeCollectionPath, handler.CreateCategory)
	routes.PATCH(routeIDPath, handler.UpdateCategory)
	routes.PATCH(routeIDPath+"/wallets", handler.UpdateCategoryWallets)
	routes.PATCH(routeIDPath+"/status", handler.SetCategoryStatus)
	routes.DELETE(routeIDPath, handler.DeleteCategory)
}
