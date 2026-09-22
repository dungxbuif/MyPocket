package httpapi

func (r *Router) RegisterFeedbackRoutes(feedback *FeedbackHandler, changelog *ChangelogHandler, agent *FeedbackAgentMiddleware) {
	publicChangelog := r.Engine.Group("/api/v1/changelog")
	publicChangelog.GET("", changelog.List)
	publicChangelog.GET("/:id", changelog.Get)

	userFeedback := r.Engine.Group("/api/v1/feedback")
	userFeedback.Use(r.AuthMiddleware.RequireAuth)
	userFeedback.GET("", feedback.List)
	userFeedback.POST("", feedback.Create)
	userFeedback.GET("/:id", feedback.Get)
	userFeedback.GET("/:id/screenshot", feedback.Screenshot)

	agentFeedback := r.Engine.Group("/api/v1/agent")
	agentFeedback.Use(agent.RequireFeedbackAgent)
	agentFeedback.GET("/feedback", feedback.AgentList)
	agentFeedback.GET("/feedback/:id/screenshot", feedback.AgentScreenshot)

	internal := r.Engine.Group("/api/v1/internal")
	internal.Use(agent.RequireFeedbackAgent)
	internal.PATCH("/feedback/:id/status", feedback.SetStatus)
	internal.POST("/changelog", changelog.Publish)
}
