package httpapi

func (r *Router) RegisterFeedbackRoutes(feedback *FeedbackHandler, changelog *ChangelogHandler, agent *FeedbackAgentMiddleware) {
	publicChangelog := r.Engine.Group("/api/v1/changelog")
	publicChangelog.GET("", changelog.List)
	publicChangelog.GET("/:id", changelog.Get)

	userFeedback := r.Engine.Group("/api/v1/feedback")
	userFeedback.GET("", r.AuthMiddleware.RequireFeedbackReadAuth, feedback.List)
	userFeedback.POST("", r.AuthMiddleware.RequireFeedbackWriteAuth, feedback.Create)
	userFeedback.GET("/:id", r.AuthMiddleware.RequireFeedbackReadAuth, feedback.Get)
	userFeedback.GET("/:id/screenshot", r.AuthMiddleware.RequireFeedbackReadAuth, feedback.Screenshot)

	agentFeedback := r.Engine.Group("/api/v1/agent")
	agentFeedback.Use(agent.RequireFeedbackAgent)
	agentFeedback.GET("/feedback", feedback.AgentList)
	agentFeedback.GET("/feedback/:id/screenshot", feedback.AgentScreenshot)

	internal := r.Engine.Group("/api/v1/internal")
	internal.Use(agent.RequireFeedbackAgent)
	internal.PATCH("/feedback/:id/status", feedback.SetStatus)
	internal.POST("/changelog", changelog.Publish)
}
