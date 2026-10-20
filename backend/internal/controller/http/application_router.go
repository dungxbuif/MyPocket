package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApplicationHandlers supplies dependencies without repeating route registration
// in the executable. Empty service values still register their routes so route
// contract verification never depends on a live database or provider.
type ApplicationHandlers struct {
	Auth          *AuthHandler
	Profile       *ProfileHandler
	Home          *HomeHandler
	Category      *CategoryHandler
	Wallet        *WalletHandler
	Transaction   *TransactionHandler
	Middleware    *AuthMiddleware
	Notification  *NotificationHandler
	APIKey        *APIKeyHandler
	Budget        *BudgetHandler
	BudgetSeries  *BudgetSeriesHandler
	BudgetReport  *BudgetReportHandler
	Savings       *SavingsReportHandler
	Jar           *JarHandler
	Month         *MonthHandler
	Recurring     *RecurringHandler
	Credit        *CreditHandler
	Travel        *TravelHandler
	Portfolio     *PortfolioHandler
	Insider       *InsiderHandler
	AIEntry       *AIEntryHandler
	Advisor       *AdvisorHandler
	Feedback      *FeedbackHandler
	Changelog     *ChangelogHandler
	FeedbackAgent *FeedbackAgentMiddleware
}

type ApplicationRouterOptions struct {
	AllowedOrigins []string
	MCPOptions     MCPHandlerOptions
	DocsHandler    gin.HandlerFunc
}

// NewApplicationRouter is the sole application route assembly used by main and
// the parity gate. Add route families here, never to an entry-point-only copy:
// every mounted method/path must then pass catalog/exclusion/OpenAPI checks.
func NewApplicationRouter(handlers ApplicationHandlers, options ApplicationRouterOptions) *Router {
	middleware := applicationHandler(handlers.Middleware)
	r := NewRouter(applicationHandler(handlers.Auth), applicationHandler(handlers.Profile), applicationHandler(handlers.Home), applicationHandler(handlers.Category), applicationHandler(handlers.Wallet), applicationHandler(handlers.Transaction), middleware, options.AllowedOrigins)
	r.RegisterNotificationRoutes(applicationHandler(handlers.Notification))
	r.RegisterAPIKeyRoutes(applicationHandler(handlers.APIKey))
	r.RegisterBudgetRoutes(applicationHandler(handlers.Budget))
	r.RegisterBudgetSeriesRoutes(applicationHandler(handlers.BudgetSeries))
	r.RegisterBudgetReportRoutes(applicationHandler(handlers.BudgetReport))
	savings := applicationHandler(handlers.Savings)
	reports := r.Engine.Group("/api/v1/reports", middleware.RequireAuth)
	reports.GET("/savings", savings.GetSavingsReport)
	reports.GET("/savings/transactions", savings.ListSavingsTransactions)
	r.RegisterJarRoutes(applicationHandler(handlers.Jar))
	r.RegisterMonthRoutes(applicationHandler(handlers.Month))
	r.RegisterRecurringRoutes(applicationHandler(handlers.Recurring))
	r.RegisterCreditRoutes(applicationHandler(handlers.Credit))
	r.RegisterTravelRoutes(applicationHandler(handlers.Travel))
	r.RegisterPortfolioRoutes(applicationHandler(handlers.Portfolio))
	r.RegisterInsiderRoutes(applicationHandler(handlers.Insider))
	r.RegisterAIEntryRoutes(applicationHandler(handlers.AIEntry))
	r.RegisterAdvisorRoutes(applicationHandler(handlers.Advisor))
	r.RegisterFeedbackRoutes(mcpFeedbackHandler(applicationHandler(handlers.Feedback)), applicationHandler(handlers.Changelog), applicationHandler(handlers.FeedbackAgent))
	docs := options.DocsHandler
	if docs == nil {
		docs = gin.WrapH(http.NotFoundHandler())
	}
	r.Engine.GET("/api/v1/docs/*any", docs)
	RegisterPublicDocsRedirect(r.Engine)
	mcpOptions := options.MCPOptions
	if mcpOptions.AllowedOrigins == nil {
		mcpOptions.AllowedOrigins = options.AllowedOrigins
	}
	if mcpOptions.Audit == nil {
		mcpOptions.Audit = middleware.Audit
	}
	mcpHandler := MCPAuthenticatedHandler(NewMCPHandler(DefaultExternalOperationCatalog(), r.Engine, mcpOptions))
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		r.Engine.Handle(method, "/api/v1/mcp", middleware.RequireAuth, mcpHandler)
	}
	return r
}

func applicationHandler[T any](handler *T) *T {
	if handler == nil {
		return new(T)
	}
	return handler
}
