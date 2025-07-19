package routes

import (
	"taskflow-api/internal/config"
	"taskflow-api/internal/handlers"
	"taskflow-api/internal/middleware"
	"taskflow-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(models.SuccessResponse{
			Message: "TaskFlow API is running",
			Data: fiber.Map{
				"status":  "healthy",
				"version": "1.0.0",
			},
		})
	})

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	projectHandler := handlers.NewProjectHandler(db)
	taskHandler := handlers.NewTaskHandler(db)

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", userHandler.CreateUser)
	auth.Post("/login", LoginHandler(db, cfg))

	// Protected routes
	protected := api.Use(middleware.JWTMiddleware(cfg))

	// User routes
	users := protected.Group("/users")
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Project routes
	projects := protected.Group("/projects")
	projects.Post("/", projectHandler.CreateProject)
	projects.Get("/", projectHandler.GetProjects)
	projects.Get("/:id", projectHandler.GetProject)
	projects.Put("/:id", projectHandler.UpdateProject)
	projects.Delete("/:id", projectHandler.DeleteProject)

	// Task routes
	tasks := protected.Group("/tasks")
	tasks.Get("/:id", taskHandler.GetTask)
	tasks.Put("/:id", taskHandler.UpdateTask)
	tasks.Delete("/:id", taskHandler.DeleteTask)
	tasks.Patch("/:id/status", taskHandler.UpdateTaskStatus)

	// Project-specific task routes
	projectTasks := protected.Group("/projects/:project_id/tasks")
	projectTasks.Post("/", taskHandler.CreateTask)
	projectTasks.Get("/", taskHandler.GetProjectTasks)
}

// LoginHandler handles user authentication
func LoginHandler(db *gorm.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid request body",
				Code:    fiber.StatusBadRequest,
			})
		}

		// Find user by email
		var user models.User
		if err := db.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid credentials",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid credentials",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Generate JWT token
		token, err := middleware.GenerateJWT(&user, cfg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to generate token",
				Code:    fiber.StatusInternalServerError,
			})
		}

		return c.JSON(models.SuccessResponse{
			Message: "Login successful",
			Data: fiber.Map{
				"token": token,
				"user":  user.ToResponse(),
			},
		})
	}
}
