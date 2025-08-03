package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tasker/internal/config"
	"tasker/internal/database"
	"tasker/internal/handler"
	"tasker/internal/middleware"
	"tasker/internal/service"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	cfg := config.MustLoad()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Создаем строку подключения к БД
	dbURL := config.BuildDBConnectionString(cfg.DB)
	dbPool, err := database.NewPool(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer dbPool.Close()

	// Инициализация сервисов
	authService := service.NewAuthService(dbPool, cfg.JWTSecret)
	taskService := service.NewTaskService(dbPool)
	userService := service.NewUserService(dbPool)
	dashboardService := service.NewDashboardService(dbPool)

	app := fiber.New()
	//healthchek
	app.Get("/health", func(c fiber.Ctx) error {
		if err := dbPool.Ping(context.Background()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).SendString("DB not ready")
		}
		return c.SendString("OK")
	})

	// Middleware
	app.Use(middleware.SlogLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
	}))

	// Инициализация обработчиков
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	userHandler := handler.NewUserHandler(userService)
	dashboardsHandler := handler.NewDashboardsHandler(dashboardService)

	// Регистрация маршрутов
	authHandler.RegisterRoutes(app)
	//app.Use(middleware.AuthMiddleware(authService))
	taskHandler.RegisterRoutes(app)
	userHandler.RegisterPublicRoutes(app)
	dashboardsHandler.RegisterRoutes(app)

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("Server failed", "error", err)
			shutdown <- syscall.SIGTERM
		}
	}()

	<-shutdown
	slog.Info("Shutting down server...")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	slog.Info("Server stopped")
}
