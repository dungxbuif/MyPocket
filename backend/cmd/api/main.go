package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	_ "github.com/mypocket/backend/docs"
	"github.com/mypocket/backend/internal/config"
	httpapi "github.com/mypocket/backend/internal/controller/http"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	"github.com/mypocket/backend/internal/infrastructure/cache"
	"github.com/mypocket/backend/internal/infrastructure/db"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title MyPocket API
// @version 0.1.0
// @description Current implemented API surface for local development.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
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

	if err := db.EnsureSchema(database); err != nil {
		log.Fatalf("migrate schema failed: %v", err)
	}

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
	verifySession := func(sessionID string) (string, error) {
		return authUc.VerifySession(context.Background(), sessionID)
	}
	middleware := httpapi.NewAuthMiddleware(jwtSvc, verifySession)
	router := httpapi.NewRouter(authHandler, profileHandler, homeHandler, middleware, cfg.CORSAllowedOrigins)
	router.Engine.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Engine.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
