package main

import (
	"log"
	"os"
	"path/filepath"

	"fleetify/config"
	"fleetify/database"
	"fleetify/handlers"
	"fleetify/middleware"
	"fleetify/models"
	"fleetify/repository"
	"fleetify/seeder"
	"fleetify/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if cfg.RunSeeder {
		seeder.Run(db)
	}

	userRepo := repository.NewUserRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	itemRepo := repository.NewMasterItemRepository(db)
	reportRepo := repository.NewReportRepository(db)
	webhookSvc := service.NewWebhookService(cfg.WebhookURL)

	h := handlers.NewHandler(userRepo, vehicleRepo, itemRepo, reportRepo, webhookSvc, cfg.UploadDir)

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})
	app.Use(logger.New())
	app.Use(cors.New())

	uploadAbs, _ := filepath.Abs(cfg.UploadDir)
	app.Static("/uploads", uploadAbs)

	frontendPath := filepath.Join("..", "frontend")
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		frontendPath = "frontend"
	}
	app.Static("/", frontendPath)

	api := app.Group("/api")
	api.Get("/health", h.Health)
	api.Get("/users", h.ListUsers)
	api.Get("/vehicles", h.ListVehicles)
	api.Get("/master-items", h.ListMasterItems)
	api.Get("/reports", middleware.RequireUser(userRepo), h.ListReports)
	api.Get("/reports/:id", middleware.RequireUser(userRepo), h.GetReport)

	auth := middleware.RequireUser(userRepo)
	saOnly := middleware.RequireRole(models.RoleSA)
	approvalOnly := middleware.RequireRole(models.RoleApproval)

	api.Post("/reports",
		auth, saOnly,
		h.CreateReport,
	)
	api.Patch("/reports/:id/approve",
		auth, approvalOnly,
		h.ApproveReport,
	)
	api.Patch("/reports/:id/complete",
		auth, saOnly,
		h.CompleteReport,
	)

	log.Printf("Fleetify running on :%s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
