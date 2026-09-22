package httpapi

import "github.com/gin-gonic/gin"

func registerTransactionRoutes(protected *gin.RouterGroup, handler *TransactionHandler) {
	routes := protected.Group(routeTransactionsPath)
	routes.GET(routeCollectionPath, handler.ListTransactions)
	routes.POST(routeCollectionPath, handler.CreateTransaction)
	routes.POST("/transfer", handler.CreateTransfer)
	routes.PATCH(routeIDPath, handler.UpdateTransaction)
	routes.DELETE(routeIDPath, handler.DeleteTransaction)
}
