package http

import (
	"context"
	"log"

	"Diplom/internal/repository"
	assetService "Diplom/internal/service/asset"
	assetVulnService "Diplom/internal/service/asset_vulnerability"
	riskService "Diplom/internal/service/risk"
	softwareService "Diplom/internal/service/software"
	threatService "Diplom/internal/service/threat"
	vulnerabilityService "Diplom/internal/service/vulnerability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(_ context.Context, db *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware: CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Middleware: Recover from panics
	app.Use(recover.New())

	// ---------- health-check ----------
	app.Get("/health", func(c *fiber.Ctx) error {
		if db != nil {
			if err := db.Ping(c.Context()); err != nil {
				log.Printf("health: db ping error: %v", err)
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"status": "degraded",
					"error":  "db not available",
				})
			}
		}

		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// ---------- DI (репозитории и сервисы) ----------

	// Справочник ПО
	softwareRepo := repository.NewSoftwareRepository(db)
	softwareSvc := softwareService.NewService(softwareRepo)
	softwareHandler := NewSoftwareHandler(softwareSvc)

	// Assets
	assetRepo := repository.NewAssetRepository(db)
	assetSvc := assetService.NewService(assetRepo)
	assetHandler := NewAssetHandler(assetSvc, softwareSvc)

	// Threats
	threatRepo := repository.NewThreatRepository(db)
	threatSvc := threatService.NewService(threatRepo)
	threatHandler := NewThreatHandler(threatSvc)

	// Vulnerabilities
	vulnRepo := repository.NewVulnerabilityRepository(db)
	vulnSvc := vulnerabilityService.NewService(vulnRepo)
	vulnHandler := NewVulnerabilityHandler(vulnSvc)

	// Asset ↔ Vulnerability links
	assetVulnRepo := repository.NewAssetVulnerabilityRepository(db)
	assetVulnSvc := assetVulnService.NewService(assetVulnRepo)
	assetVulnHandler := NewAssetVulnerabilityHandler(assetVulnSvc)

	// Risk service (preview only, без сохранения в БД)
	riskSvc := riskService.NewService(assetRepo, threatRepo, vulnRepo, assetVulnRepo)
	riskHandler := NewRiskHandler(riskSvc)

	// ---------- Роуты ----------
	api := app.Group("/api")

	assetsGroup := api.Group("/assets")
	assetHandler.Register(assetsGroup)
	assetVulnHandler.Register(assetsGroup)

	threatHandler.Register(api.Group("/threats"))
	vulnHandler.Register(api.Group("/vulnerabilities"))
	// /api/risk/preview
	riskHandler.Register(api.Group("/risk"))
	// /api/software — справочник ПО
	softwareHandler.Register(api.Group("/software"))

	// Дополнительные алиасы для рекомендаций по ПО
	registerSoftwareAliases(api, softwareSvc)

	return app
}

func registerSoftwareAliases(api fiber.Router, svc softwareService.Service) {
	api.Get("/assets/:id/software/alternatives", func(c *fiber.Ctx) error {
		assetID, err := c.ParamsInt("id")
		if err != nil || assetID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset id"})
		}

		data, err := svc.SuggestAlternativesForAsset(c.Context(), int64(assetID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(data)
	})

	api.Get("/software/asset/:assetID/alternatives", func(c *fiber.Ctx) error {
		assetID, err := c.ParamsInt("assetID")
		if err != nil || assetID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset id"})
		}

		data, err := svc.SuggestAlternativesForAsset(c.Context(), int64(assetID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(data)
	})
}
