package httpapi

func (r *Router) RegisterNotificationRoutes(handler *NotificationHandler) {
	routes := r.Engine.Group("/api/v1/notifications")
	routes.Use(r.AuthMiddleware.RequireAuth)
	routes.GET("", handler.List)
	routes.POST("/:id/read", handler.MarkRead)
}
