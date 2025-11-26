package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
	"github.com/rcdevgames/modular-monolith-clean/internal/infrastructure/database"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/healthcheck"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/httpresp"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/logging"
	servermiddleware "github.com/rcdevgames/modular-monolith-clean/internal/server/middleware"
	docs "github.com/rcdevgames/modular-monolith-clean/internal/server/swagger/docs"
	"github.com/rcdevgames/modular-monolith-clean/internal/storage"
	"go.uber.org/zap"
)

// @title Modular Monolith Clean API
// @version 1.0
// @description API documentation for the Modular Monolith Clean server.
// @BasePath /api/v1

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	store, err := storage.New(cfg.Storage)
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}

	cleanup, err := logging.Setup(cfg.App.LogDir)
	if err != nil {
		log.Fatalf("logger setup: %v", err)
	}
	defer cleanup()

	logger := logging.GetLogger()

	ctr := container.New(cfg, db, store)

	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP, servermiddleware.HTTPLogger(), middleware.Recoverer)
	router.Use(servermiddleware.SecurityHeaders())
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpresp.Error(w, http.StatusNotFound, "Route not found", nil)
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpresp.Error(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	})

	// CDN route for serving static files
	router.Get("/cdn/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/cdn/", http.FileServer(http.Dir(cfg.Storage.LocalDir)))
		fs.ServeHTTP(w, r)
	})

	healthHandler := healthcheck.NewHandler(db)
	healthHandler.RegisterRoutes(router)

	apiRouter := chi.NewRouter()
	apiRouter.Use(servermiddleware.XSS())
	maxAgeSeconds := int(cfg.App.CORS.MaxAge / time.Second)
	if maxAgeSeconds < 0 {
		maxAgeSeconds = 0
	}
	apiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.App.CORS.AllowedOrigins,
		AllowedMethods:   cfg.App.CORS.AllowedMethods,
		AllowedHeaders:   cfg.App.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.App.CORS.ExposedHeaders,
		AllowCredentials: cfg.App.CORS.AllowCredentials,
		MaxAge:           maxAgeSeconds,
	}))
	apiRouter.Use(httprate.Limit(cfg.App.RateLimitRequests, cfg.App.RateLimitWindow))
	apiRouter.Use(servermiddleware.JWT(cfg.Security.JWTSecret, cfg.Security.JWTExpiration))
	apiRouter.Use(servermiddleware.InterService(cfg.Security.InterServiceToken))
	if err := modules.RegisterAll(ctr, apiRouter); err != nil {
		logger.Fatal("register modules", zap.Error(err))
	}

	docsBase := "/docs"
	docs.SwaggerInfo.BasePath = "/api/v1"
	if cfg.App.Env != "production" {
		router.Get(docsBase, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, docsBase+"/index.html", http.StatusTemporaryRedirect)
		})
		router.Get(docsBase+"/*", httpSwagger.Handler(
			httpSwagger.URL(docsBase+"/doc.json"),
		))
	}
	router.Mount("/api/v1", apiRouter)

	addr := ":" + cfg.App.Port
	logger.Info("API listening", zap.String("address", addr))
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
