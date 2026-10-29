package httpapi

import "github.com/gin-gonic/gin"

func registerTransactionRoutes(protected *gin.RouterGroup, handler *TransactionHandler) {
	routes := protected.Group(routeTransactionsPath)
	routes.GET(routeCollectionPath, handler.ListTransactions)
	routes.POST(routeCollectionPath, handler.CreateTransaction)
	routes.POST("/adjustment", handler.CreateAdjustment)
	routes.POST("/bulk-delete", handler.BulkDeleteTransactions)
	routes.POST("/transfer", handler.CreateTransfer)
	routes.PATCH("/transfer/:transfer_id", handler.UpdateTransfer)
	routes.DELETE("/transfer/:transfer_id", handler.DeleteTransfer)
	routes.PATCH(routeIDPath, handler.UpdateTransaction)
	routes.PATCH(routeIDPath+"/travel", handler.UpdateTransactionTravel)
	routes.DELETE(routeIDPath, handler.DeleteTransaction)
}
