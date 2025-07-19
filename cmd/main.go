package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow-api/internal/config"
	"taskflow-api/internal/models"
	"taskflow-api/internal/routes"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate database schema
	if err := autoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(models.ErrorResponse{
				Error:   "Error",
				Message: err.Error(),
				Code:    code,
			})
		},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Setup routes
	routes.SetupRoutes(app, db, cfg)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")

		// Close database connection
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}

		// Shutdown server
		app.Shutdown()
	}()

	// Start server
	log.Printf("🚀 TaskFlow API starting on port %s", cfg.Port)
	log.Printf("📝 Environment: %s", cfg.Environment)
	log.Printf("🏥 Health check: http://localhost:%s/health", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if cfg.Environment == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("✅ Database connected successfully")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Task{},
	); err != nil {
		return err
	}

	log.Println("✅ Database migrations completed")
	return nil
}
