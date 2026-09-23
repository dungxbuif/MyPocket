package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

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
	"github.com/mypocket/backend/internal/infrastructure/storage"
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
	jarRepository := repo.NewJarPostgresRepository(database)
	monthNotes := repo.NewMonthPostgresRepository(database)
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
	transactionRepository := repo.NewTransactionPostgresRepository(database)
	transactionHandler := httpapi.NewTransactionHandler(transactionRepository, walletRepository, categoryRepository)
	transactionHandler.Users = userRepo
	transactionHandler.Jars = jarRepository
	verifySession := func(sessionID string) (string, error) {
		return authUc.VerifySession(context.Background(), sessionID)
	}
	middleware := httpapi.NewAuthMiddleware(jwtSvc, verifySession)
	apiKeyService := usecase.NewUserAPIKeyService(repo.NewUserAPIKeyPostgresRepository(database))
	middleware.APIKeys = apiKeyService
	middleware.Audit = cacheRepo
	router := httpapi.NewRouter(authHandler, profileHandler, homeHandler, categoryHandler, walletHandler, transactionHandler, middleware, cfg.CORSAllowedOrigins)
	router.RegisterAPIKeyRoutes(&httpapi.APIKeyHandler{Service: apiKeyService})
	router.RegisterBudgetRoutes(&httpapi.BudgetHandler{Budgets: repo.NewBudgetPostgresRepository(database), Wallets: walletRepository, Categories: categoryRepository, Transactions: repo.NewTransactionPostgresRepository(database), Users: userRepo})
	router.RegisterJarRoutes(&httpapi.JarHandler{Jars: jarRepository, Users: userRepo})
	router.RegisterMonthRoutes(&httpapi.MonthHandler{Users: userRepo, Transactions: transactionRepository, Categories: categoryRepository, Jars: jarRepository, Notes: monthNotes})
	aiClient := ai.NewClient(ai.Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, OCRURL: cfg.OCRAPIURL, OCRKey: cfg.OCRAPIKey, StoreUsage: cfg.AIStoreUsage})
	attachmentStorage, err := newAttachmentStorage(cfg)
	if err != nil {
		log.Fatalf("configure private attachment storage failed: %v", err)
	}
	router.RegisterAIEntryRoutes(&httpapi.AIEntryHandler{Service: &usecase.AIEntryService{Entries: repo.NewAIEntryPostgresRepository(database), Wallets: walletRepository, Categories: categoryRepository, Extractor: aiClient, Storage: attachmentStorage, Users: userRepo}})
	financeReader := repo.NewFinanceQueryPostgresRepository(database)
	financeQueryService := usecase.NewFinanceQueryService(financeReader)
	advisorProvider := ai.NewAdvisorClient(ai.AdvisorClientConfig{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, StoreUsage: cfg.AIStoreUsage})
	advisorStore := repo.NewAdvisorPostgresRepository(database)
	advisorOrchestrator := usecase.NewAdvisorOrchestrator(advisorProvider, usecase.NewAdvisorToolRegistry(financeQueryService))
	advisorOrchestrator.Validator = usecase.AdvisorPrincipalValidatorFunc(func(ctx context.Context, principal usecase.Principal) error {
		switch principal.CredentialKind {
		case "user_api_key":
			return apiKeyService.ValidatePrincipal(ctx, principal)
		case "session":
			if !principal.ExpiresAt.IsZero() && !principal.ExpiresAt.After(time.Now().UTC()) {
				return usecase.ErrAdvisorPrincipalInvalid
			}
			ownerID, err := verifySession(principal.CredentialID)
			if err != nil || ownerID != principal.OwnerID {
				return usecase.ErrAdvisorPrincipalInvalid
			}
			return nil
		default:
			return usecase.ErrAdvisorPrincipalInvalid
		}
	})
	advisorService := usecase.NewAdvisorService(advisorStore, advisorOrchestrator)
	router.RegisterAdvisorRoutes(&httpapi.AdvisorHandler{Service: advisorService, Users: userRepo, Finance: financeQueryService})
	feedbackRepository := repo.NewFeedbackPostgresRepository(database)
	changelogRepository := repo.NewChangelogPostgresRepository(database)
	feedbackService := usecase.NewFeedbackServiceWithStorage(feedbackRepository, cacheRepo, attachmentStorage, time.Now)
	changelogService := usecase.NewChangelogService(feedbackRepository, changelogRepository, cacheRepo, time.Now)
	if cfg.AppEnv != "development" && strings.TrimSpace(cfg.FeedbackAgentToken) == "" {
		log.Fatal("FEEDBACK_AGENT_TOKEN must be configured outside development")
	}
	router.RegisterFeedbackRoutes(&httpapi.FeedbackHandler{Service: feedbackService}, &httpapi.ChangelogHandler{Service: changelogService}, httpapi.NewFeedbackAgentMiddleware(cfg.FeedbackAgentToken))
	router.Engine.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func newAttachmentStorage(cfg config.Config) (*storage.S3, error) {
	configured := cfg.S3Endpoint != "" || cfg.S3Region != "" || cfg.S3Bucket != "" || cfg.S3AccessKeyID != "" || cfg.S3SecretAccessKey != ""
	if !configured {
		if cfg.AppEnv != "development" {
			return nil, errors.New("private attachment storage must be configured outside development")
		}
		return nil, nil
	}
	return storage.NewS3(storage.S3Config{Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket, Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle})
}
