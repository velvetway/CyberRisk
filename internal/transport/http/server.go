package http

import (
	"context"
	"log"

	"Diplom/internal/repository"
	assetService "Diplom/internal/service/asset"
	assetVulnService "Diplom/internal/service/asset_vulnerability"
	authService "Diplom/internal/service/auth"
	riskService "Diplom/internal/service/risk"
	softwareService "Diplom/internal/service/software"
	threatService "Diplom/internal/service/threat"
	vulnerabilityService "Diplom/internal/service/vulnerability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(_ context.Context, db *pgxpool.Pool, jwtSecret string) *fiber.App {
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
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ---------- DI ----------

	// Auth
	userRepo := repository.NewUserRepository(db)
	authSvc := authService.NewService(userRepo, jwtSecret)
	authHandler := NewAuthHandler(authSvc)

	// Software
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

	// Asset-Vulnerability links
	assetVulnRepo := repository.NewAssetVulnerabilityRepository(db)
	assetVulnSvc := assetVulnService.NewService(assetVulnRepo)
	assetVulnHandler := NewAssetVulnerabilityHandler(assetVulnSvc)

	// Risk
	riskSvc := riskService.NewService(assetRepo, threatRepo, vulnRepo, assetVulnRepo)
	riskHandler := NewRiskHandler(riskSvc)

	// ---------- Public routes (no auth) ----------
	api := app.Group("/api")

	authGroup := api.Group("/auth")
	authHandler.RegisterRoutes(authGroup) // POST /api/auth/register, POST /api/auth/login

	// ---------- Protected routes (JWT required) ----------
	protected := api.Group("", JWTMiddleware(authSvc))

	// GET /api/auth/me
	authHandler.RegisterProtected(protected.Group("/auth"))

	// Role groups
	readOnly := protected.Group("", RequireRole("viewer", "auditor", "admin"))
	write := protected.Group("", RequireRole("auditor", "admin"))
	adminOnly := protected.Group("", RequireRole("admin"))

	// Assets
	assetsRead := readOnly.Group("/assets")
	assetsRead.Get("/", assetHandler.listAssets)
	assetsRead.Get("/:id", assetHandler.getAsset)
	assetsRead.Get("/:assetID/vulnerabilities", assetVulnHandler.listForAsset)
	assetsRead.Get("/:id/software/alternatives", assetHandler.assetSoftwareAlternatives)

	assetsWrite := write.Group("/assets")
	assetsWrite.Post("/", assetHandler.createAsset)
	assetsWrite.Put("/:id", assetHandler.updateAsset)
	assetsWrite.Post("/:assetID/vulnerabilities", assetVulnHandler.addToAsset)

	assetsAdmin := adminOnly.Group("/assets")
	assetsAdmin.Delete("/:id", assetHandler.deleteAsset)
	assetsAdmin.Delete("/:assetID/vulnerabilities/:vulnID", assetVulnHandler.removeFromAsset)

	// Threats
	readOnly.Get("/threats", threatHandler.listThreats)
	readOnly.Get("/threats/:id", threatHandler.getThreat)
	write.Post("/threats", threatHandler.createThreat)
	write.Put("/threats/:id", threatHandler.updateThreat)
	adminOnly.Delete("/threats/:id", threatHandler.deleteThreat)

	// Vulnerabilities
	readOnly.Get("/vulnerabilities", vulnHandler.list)
	readOnly.Get("/vulnerabilities/:id", vulnHandler.get)
	write.Post("/vulnerabilities", vulnHandler.create)
	write.Put("/vulnerabilities/:id", vulnHandler.update)
	adminOnly.Delete("/vulnerabilities/:id", vulnHandler.delete)

	// Risk
	readOnly.Get("/risk/overview", riskHandler.overview)
	readOnly.Get("/risk/asset/:id", riskHandler.assetRiskProfile)
	write.Post("/risk/preview", riskHandler.previewRisk)
	write.Post("/risk/report/pdf", riskHandler.GenerateRiskPDF)

	// Software
	readOnly.Get("/software", softwareHandler.listSoftware)
	readOnly.Get("/software/categories", softwareHandler.listCategories)
	readOnly.Get("/software/russian", softwareHandler.listRussianSoftware)
	readOnly.Get("/software/certified", softwareHandler.listCertifiedSoftware)
	readOnly.Get("/software/:id", softwareHandler.getSoftware)
	write.Post("/software", softwareHandler.createSoftware)
	write.Put("/software/:id", softwareHandler.updateSoftware)
	adminOnly.Delete("/software/:id", softwareHandler.deleteSoftware)

	// Software aliases
	readOnly.Get("/software/asset/:assetID/alternatives", softwareHandler.assetAlternatives)

	return app
}
