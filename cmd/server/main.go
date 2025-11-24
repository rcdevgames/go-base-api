package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
	"github.com/rcdevgames/modular-monolith-clean/internal/infrastructure/database"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/healthcheck"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/httpresp"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/logging"
	servermiddleware "github.com/rcdevgames/modular-monolith-clean/internal/server/middleware"
	"github.com/rcdevgames/modular-monolith-clean/internal/storage"
)

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

	if _, err := logging.Setup(cfg.App.LogDir); err != nil {
		log.Fatalf("logger setup: %v", err)
	}

	ctr := container.New(cfg, db, store)

	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	router.Use(servermiddleware.SecurityHeaders())
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpresp.Error(w, http.StatusNotFound, "route not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpresp.Error(w, http.StatusMethodNotAllowed, "method not allowed")
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
		log.Fatalf("register modules: %v", err)
	}

	router.Mount("/api/v1", apiRouter)

	addr := ":" + cfg.App.Port
	log.Printf("API listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
