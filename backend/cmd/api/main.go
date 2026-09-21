package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	_ "github.com/mypocket/backend/docs"
	"github.com/mypocket/backend/internal/config"
	httpapi "github.com/mypocket/backend/internal/controller/http"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/ai"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	"github.com/mypocket/backend/internal/infrastructure/cache"
	"github.com/mypocket/backend/internal/infrastructure/db"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d .,../../internal/controller/http,../../internal/usecase,../../internal/entity -o ../../docs --parseInternal

// @title MyPocket API
// @version 0.1.0
// @description Current implemented API surface for local development.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := config.LoadLocalEnv(".env.local"); err != nil {
		log.Fatal(err)
	}
	cfg := config.Load()

	database, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres failed: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("open sql database handle failed: %v", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	cacheRepo, err := cache.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("connect redis failed: %v", err)
	}
	defer func() {
		_ = cacheRepo.Close()
	}()

	passwordSvc := auth.NewPasswordService()

	hashed, err := passwordSvc.Hash(cfg.SeedPassword)
	if err != nil {
		log.Fatalf("seed hash failed: %v", err)
	}
	seedUser := &entity.User{
		ID:           uuid.NewString(),
		Name:         cfg.SeedName,
		Email:        cfg.SeedEmail,
		PasswordHash: hashed,
	}
	if err := db.SeedDefaultUser(database, seedUser); err != nil {
		log.Fatalf("seed default user failed: %v", err)
	}

	userRepo := repo.NewUserPostgresRepository(database)
	jwtSvc := auth.NewJWT(cfg.JWTSecret, cfg.JWTTTL)
	googleOAuth := auth.NewGoogleOAuth(auth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	})
	authUc := usecase.NewAuthInteractor(userRepo, passwordSvc, jwtSvc, cacheRepo, googleOAuth)

	authHandler := httpapi.NewAuthHandler(
		authUc,
		cfg.OAuthFixtureMode,
		cfg.AllowedLoginEmails,
	)
	profileHandler := httpapi.NewProfileHandler(authUc)
	homeHandler := httpapi.NewHomeHandler(authUc)
	categoryRepository := repo.NewCategoryPostgresRepository(database)
	categoryHandler := httpapi.NewCategoryHandler(usecase.NewCategoryInteractor(categoryRepository))
	walletRepository := repo.NewWalletPostgresRepository(database)
	walletHandler := httpapi.NewWalletHandler(walletRepository)
	transactionHandler := httpapi.NewTransactionHandler(repo.NewTransactionPostgresRepository(database), walletRepository, categoryRepository)
	verifySession := func(sessionID string) (string, error) {
		return authUc.VerifySession(context.Background(), sessionID)
	}
	middleware := httpapi.NewAuthMiddleware(jwtSvc, verifySession)
	router := httpapi.NewRouter(authHandler, profileHandler, homeHandler, categoryHandler, walletHandler, transactionHandler, middleware, cfg.CORSAllowedOrigins)
	router.RegisterBudgetRoutes(&httpapi.BudgetHandler{Budgets: repo.NewBudgetPostgresRepository(database), Wallets: walletRepository, Categories: categoryRepository, Transactions: repo.NewTransactionPostgresRepository(database)})
	aiClient := ai.NewClient(ai.Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, OCRURL: cfg.OCRAPIURL, OCRKey: cfg.OCRAPIKey})
	router.RegisterAIEntryRoutes(&httpapi.AIEntryHandler{Service: &usecase.AIEntryService{Entries: repo.NewAIEntryPostgresRepository(database), Wallets: walletRepository, Categories: categoryRepository, Extractor: aiClient}})
	router.Engine.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
