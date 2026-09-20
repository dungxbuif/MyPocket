package httpapi

func (r *Router) RegisterBudgetRoutes(handler *BudgetHandler) {
	routes := r.Engine.Group("/api/v1/budgets")
	routes.Use(r.AuthMiddleware.RequireAuth)
	routes.GET("", handler.ListBudgets)
	routes.POST("", handler.CreateBudget)
	routes.PATCH("/:id", handler.UpdateBudget)
	routes.DELETE("/:id", handler.DeleteBudget)
}
