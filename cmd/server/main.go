package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"ipl-be-svc/docs"
	"ipl-be-svc/internal/config"
	"ipl-be-svc/internal/database"
	"ipl-be-svc/internal/handler"
	"ipl-be-svc/internal/middleware"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
)

// @title IPL Backend Service API
// @version 1.0
// @description RESTful API for IPL Backend Service with menu management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Swagger documentation
	docs.SwaggerInfo.Title = "IPL Backend Service API"
	docs.SwaggerInfo.Description = "RESTful API for IPL Backend Service with menu management"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Server.Port)
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Initialize logger
	appLogger := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format)
	appLogger.Info("Starting IPL Backend Service...")

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		appLogger.WithField("error", err).Fatal("Failed to connect to database")
	}
	appLogger.Info("Database connected successfully")

	// Run auto migration
	if err := db.AutoMigrate(); err != nil {
		appLogger.WithField("error", err).Fatal("Failed to run database migrations")
	}
	appLogger.Info("Database migrations completed successfully")

	// Initialize repositories
	menuRepo := repository.NewMenuRepository(db.DB)
	billingRepo := repository.NewBillingRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	masterMenuRepo := repository.NewMasterMenuRepository(db.DB)
	roleMenuRepo := repository.NewRoleMenuRepository(db.DB)
	dashboardRepo := repository.NewDashboardRepository(db.DB)

	// Initialize services
	menuService := service.NewMenuService(menuRepo)
	dokuService := service.NewDokuService(appLogger)
	paymentService := service.NewPaymentService(billingRepo, dokuService, appLogger)
	userService := service.NewUserService(userRepo, appLogger)
	billingService := service.NewBillingService(billingRepo, db.DB)
	masterMenuService := service.NewMasterMenuService(masterMenuRepo, appLogger)
	roleMenuService := service.NewRoleMenuService(roleMenuRepo, masterMenuRepo, appLogger)
	dashboardService := service.NewDashboardService(dashboardRepo, appLogger)

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.LoggerMiddleware(appLogger))
	router.Use(middleware.ErrorHandler())
	router.NoRoute(middleware.NoRouteHandler())
	router.NoMethod(middleware.NoMethodHandler())

	// Setup routes
	handler.SetupRoutes(router, menuService, paymentService, userService, billingService, masterMenuService, roleMenuService, dashboardService, appLogger)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithField("port", cfg.Server.Port).Info("Server starting...")
		appLogger.WithField("swagger", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.Server.Port)).Info("Swagger documentation available")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	appLogger.WithField("port", cfg.Server.Port).Info("Server started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithField("error", err).Fatal("Server forced to shutdown")
	}

	// Close database connection
	if err := db.Close(); err != nil {
		appLogger.WithField("error", err).Error("Failed to close database connection")
	}

	appLogger.Info("Server exited successfully")
}
