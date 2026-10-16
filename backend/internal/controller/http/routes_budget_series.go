package httpapi

func (r *Router) RegisterBudgetSeriesRoutes(handler *BudgetSeriesHandler) {
	routes := r.Engine.Group("/api/v1/budget-series")
	routes.Use(r.AuthMiddleware.RequireAuth)
	routes.GET("", handler.List)
	routes.POST("", handler.Create)
	routes.PATCH("/:id", handler.Update)
	routes.POST("/:id/stop", handler.Stop)
}
