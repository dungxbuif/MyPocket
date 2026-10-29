package httpapi

import "github.com/gin-gonic/gin"

func registerWalletRoutes(protected *gin.RouterGroup, handler *WalletHandler) {
	routes := protected.Group(routeWalletsPath)
	routes.GET(routeCollectionPath, handler.ListWallets)
	routes.POST(routeCollectionPath, handler.CreateWallet)
	routes.PATCH(routeIDPath, handler.UpdateWallet)
	routes.DELETE(routeIDPath, handler.DeleteWallet)
}
