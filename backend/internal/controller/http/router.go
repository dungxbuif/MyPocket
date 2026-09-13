package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsHeaderAllowOrigin      = "Access-Control-Allow-Origin"
	corsHeaderAllowCredentials = "Access-Control-Allow-Credentials"
	corsHeaderAllowHeaders     = "Access-Control-Allow-Headers"
	corsHeaderAllowMethods     = "Access-Control-Allow-Methods"
	corsHeaderVary             = "Vary"
	corsAllowedHeaders         = "Authorization, Content-Type"
	corsAllowedMethods         = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
)

type Router struct {
	Engine             *gin.Engine
	AuthHandler        *AuthHandler
	ProfileHandler     *ProfileHandler
	HomeHandler        *HomeHandler
	CategoryHandler    *CategoryHandler
	WalletHandler      *WalletHandler
	TransactionHandler *TransactionHandler
	AuthMiddleware     *AuthMiddleware
}

func NewRouter(authHandler *AuthHandler, profileHandler *ProfileHandler, homeHandler *HomeHandler, categoryHandler *CategoryHandler, walletHandler *WalletHandler, transactionHandler *TransactionHandler, middleware *AuthMiddleware, allowedOrigins []string) *Router {
	engine := gin.Default()
	engine.Use(requestIDMiddleware())
	engine.Use(corsMiddleware(allowedOrigins))
	r := &Router{
		Engine:             engine,
		AuthHandler:        authHandler,
		ProfileHandler:     profileHandler,
		HomeHandler:        homeHandler,
		CategoryHandler:    categoryHandler,
		WalletHandler:      walletHandler,
		TransactionHandler: transactionHandler,
		AuthMiddleware:     middleware,
	}
	api := engine.Group(apiV1Path)
	registerPublicRoutes(api, authHandler)
	protected := api.Group(routeRootPath)
	protected.Use(middleware.RequireAuth)
	registerAccountRoutes(protected, profileHandler, homeHandler)
	registerCategoryRoutes(protected, categoryHandler)
	registerWalletRoutes(protected, walletHandler)
	registerTransactionRoutes(protected, transactionHandler)
	return r
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if normalized := strings.TrimSpace(origin); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if _, ok := allowed[origin]; ok {
			c.Header(corsHeaderAllowOrigin, origin)
			c.Header(corsHeaderAllowCredentials, "true")
			c.Header(corsHeaderAllowHeaders, corsAllowedHeaders)
			c.Header(corsHeaderAllowMethods, corsAllowedMethods)
			c.Header(corsHeaderVary, "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
