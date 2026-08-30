package http

import (
	"context"
	"log"

	"Diplom/internal/bdu"
	"Diplom/internal/config"
	"Diplom/internal/repository"
	assetService "Diplom/internal/service/asset"
	assetVulnService "Diplom/internal/service/asset_vulnerability"
	authService "Diplom/internal/service/auth"
	complianceService "Diplom/internal/service/compliance"
	controlService "Diplom/internal/service/control"
	externalCatalogService "Diplom/internal/service/external_catalog"
	optimizerService "Diplom/internal/service/optimizer"
	organizationService "Diplom/internal/service/organization"
	ptsziService "Diplom/internal/service/ptszi"
	riskService "Diplom/internal/service/risk"
	softwareService "Diplom/internal/service/software"
	threatService "Diplom/internal/service/threat"
	vulnerabilityService "Diplom/internal/service/vulnerability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

// detectorBinder lets us call the optional SetVulnerabilityDetector method
// on softwareService.service without exposing the concrete struct.
type detectorBinder interface {
	SetVulnerabilityDetector(softwareService.VulnerabilityDetector)
}

// NewServer принимает конфиг целиком: путей к локальным каталогам стало
// столько, что перечислять их по одному в сигнатуре уже неудобно.
func NewServer(_ context.Context, db *pgxpool.Pool, cfg *config.Config, bduSnap *bdu.Snapshot) *fiber.App {
	jwtSecret := cfg.JWTSecret
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

	// Asset-Vulnerability links + autodetect через bdu-snapshot
	assetVulnRepo := repository.NewAssetVulnerabilityRepository(db)
	vlReader := repository.NewVLCategoryReader(db)
	assetVulnSvc := assetVulnService.NewService(assetVulnRepo, softwareRepo, bduSnap, vlReader)
	assetVulnHandler := NewAssetVulnerabilityHandler(assetVulnSvc)

	// Wire autodetect into software service so AttachToAsset/DetachFromAsset
	// auto-populate / cleanup asset_vulnerabilities for that ПО.
	if d, ok := softwareSvc.(detectorBinder); ok {
		d.SetVulnerabilityDetector(assetVulnSvc)
	}

	// Локальные каталоги: БДУ ФСТЭК, Минреестр и реестр сертифицированных СЗИ.
	//
	// БДУ читается двумя способами и это намеренно: bduSnap выше отвечает за
	// автодетект уязвимостей конкретной версии ПО, bduRepo — за поиск по
	// каталогу с фильтрами. Задачи разные, файл один.
	bduRepo, err := repository.NewBDUSQLiteRepository(cfg.BDUSQLitePath)
	if err != nil {
		log.Printf("bdu sqlite disabled: %v", err)
	}
	minreestrRepo, err := repository.NewMinreestrSQLiteRepository(cfg.MinreestrSQLitePath)
	if err != nil {
		log.Printf("minreestr sqlite disabled: %v", err)
	}
	sziRepo, err := repository.NewSZISQLiteRepository(cfg.SZISQLitePath)
	if err != nil {
		log.Printf("szi sqlite disabled: %v", err)
	}
	externalCatalogSvc := externalCatalogService.NewService(bduRepo, minreestrRepo, sziRepo)
	externalCatalogHandler := NewExternalCatalogHandler(externalCatalogSvc)

	// Canonical PTSZI model: S -> ST -> VL -> controls -> DA -> W.
	ptsziRepo := repository.NewPTSZIRepository(db)
	ptsziSvc := ptsziService.NewService(assetRepo, ptsziRepo)
	ptsziHandler := NewPTSZIHandler(ptsziSvc)

	// Подбор комплекса СЗИ: снижение суммарного W при ограниченном бюджете.
	optimizerSvc := optimizerService.NewService(ptsziSvc, sziRepo)
	optimizerHandler := NewOptimizerHandler(optimizerSvc)

	// Controls (catalogue + asset_controls bridge)
	controlRepo := repository.NewControlRepository(db)
	controlSvc := controlService.NewService(controlRepo)
	controlHandler := NewControlHandler(controlSvc)

	// Risk (PTSZI W-model)
	threatSourceRepo := repository.NewThreatSourceRepository(db)
	destructiveActionRepo := repository.NewDestructiveActionRepository(db)
	riskGraphRepo := repository.NewRiskGraphRepository(db)
	riskSvc := riskService.NewService(assetRepo, threatRepo, threatSourceRepo, destructiveActionRepo, riskGraphRepo)
	riskHandler := NewRiskHandler(riskSvc)

	// Compliance (ОСЗ — оценка состояния защищённости активa по стандартам)
	complianceRepo := repository.NewComplianceRepository(db)
	complianceSvc := complianceService.NewService(complianceRepo, controlRepo, assetRepo)
	complianceHandler := NewComplianceHandler(complianceSvc, assetRepo)

	// Documents (орг.-тех. документация: паспорт АС, модель угроз, план защиты)
	assetTypeNameLookup := func(ctx context.Context, id int16) string {
		var name string
		if err := db.QueryRow(ctx, `SELECT name FROM asset_types WHERE id = $1`, id).Scan(&name); err != nil {
			return ""
		}
		return name
	}
	documentHandler := NewDocumentHandler(
		assetRepo, softwareRepo, controlRepo, assetVulnRepo,
		riskSvc, complianceSvc, assetTypeNameLookup, db,
	)

	// Organization-level overview / отчёт по всей организации
	organizationSvc := organizationService.NewService(assetRepo, controlRepo, riskSvc, complianceSvc, assetTypeNameLookup)
	organizationHandler := NewOrganizationHandler(organizationSvc)

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
	assetsRead.Get("/:assetID/software", softwareHandler.listAssetSoftware)
	assetsRead.Get("/:assetID/controls", controlHandler.listForAsset)
	assetsRead.Get("/:id/software/alternatives", assetHandler.assetSoftwareAlternatives)

	assetsWrite := write.Group("/assets")
	assetsWrite.Post("/", assetHandler.createAsset)
	assetsWrite.Put("/:id", assetHandler.updateAsset)
	assetsWrite.Post("/:assetID/vulnerabilities", assetVulnHandler.addToAsset)
	assetsWrite.Post("/:assetID/software", softwareHandler.attachAssetSoftware)
	assetsWrite.Delete("/:assetID/software/:softwareID", softwareHandler.detachAssetSoftware)
	assetsWrite.Post("/:assetID/controls", controlHandler.attachToAsset)
	assetsWrite.Delete("/:assetID/controls/:controlID", controlHandler.detachFromAsset)

	assetsAdmin := adminOnly.Group("/assets")
	assetsAdmin.Delete("/:id", assetHandler.deleteAsset)
	assetsAdmin.Delete("/:assetID/vulnerabilities/:vulnID", assetVulnHandler.removeFromAsset)

	// Catalogue: GET /api/controls — список всех методов противодействия
	readOnly.Get("/controls", controlHandler.list)

	// Asset types reference (для UI-маппинга id→имя в каталоге угроз).
	readOnly.Get("/asset-types", func(c *fiber.Ctx) error {
		rows, err := db.Query(c.Context(), `SELECT id, name, COALESCE(description, '') FROM asset_types ORDER BY id`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()
		type row struct {
			ID          int16  `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		out := make([]row, 0, 8)
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.ID, &r.Name, &r.Description); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			out = append(out, r)
		}
		return c.JSON(out)
	})

	// Data categories — категории обрабатываемой информации (для карточки актива).
	readOnly.Get("/data-categories", func(c *fiber.Ctx) error {
		rows, err := db.Query(c.Context(), `SELECT id, code, name, COALESCE(description, '') FROM data_categories ORDER BY sort_order, id`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()
		type row struct {
			ID          int16  `json:"id"`
			Code        string `json:"code"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		out := make([]row, 0, 8)
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.ID, &r.Code, &r.Name, &r.Description); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			out = append(out, r)
		}
		return c.JSON(out)
	})

	// VL-категории справочник (UI-фильтры P9, графика P8).
	readOnly.Get("/vl-categories", func(c *fiber.Ctx) error {
		rows, err := db.Query(c.Context(), `SELECT id, code, name, COALESCE(description, '') FROM vl_categories ORDER BY id`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()
		type row struct {
			ID          int16  `json:"id"`
			Code        string `json:"code"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		out := make([]row, 0, 6)
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.ID, &r.Code, &r.Name, &r.Description); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			out = append(out, r)
		}
		return c.JSON(out)
	})

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

	// External local catalogs
	readOnly.Get("/external/bdu/vulnerabilities", externalCatalogHandler.searchBDU)
	readOnly.Get("/external/minreestr/software", externalCatalogHandler.searchMinreestr)
	readOnly.Get("/external/szi/certificates", externalCatalogHandler.searchSZI)
	readOnly.Get("/external/szi/coverage", externalCatalogHandler.sziCoverage)

	// Canonical PTSZI routes
	readOnly.Get("/ptszi/sources", ptsziHandler.listSources)
	readOnly.Get("/ptszi/threats", ptsziHandler.listThreats)
	readOnly.Get("/ptszi/vulnerable-links", ptsziHandler.listVulnerableLinks)
	readOnly.Get("/ptszi/controls", ptsziHandler.listControls)
	readOnly.Get("/ptszi/destructive-actions", ptsziHandler.listDestructiveActions)
	readOnly.Get("/ptszi/ubi", ptsziHandler.listUBI)
	readOnly.Get("/assets/:id/ptszi/profile", ptsziHandler.assetProfile)
	readOnly.Get("/ptszi/assets/:assetID/threats", ptsziHandler.applicableThreats)
	readOnly.Get("/ptszi/graph/:assetID/:threatID", ptsziHandler.graph)
	readOnly.Get("/ptszi/assets/:assetID/optimize", optimizerHandler.optimize)
	readOnly.Get("/ptszi/assets/:assetID/roadmap", optimizerHandler.roadmap)
	readOnly.Get("/ptszi/assets/:assetID/sensitivity", optimizerHandler.sensitivity)
	write.Put("/assets/:id/ptszi/vulnerable-links", ptsziHandler.updateAssetVulnerableLinks)
	write.Put("/assets/:id/ptszi/controls", ptsziHandler.updateAssetControls)

	// Risk (PTSZI W-model)
	readOnly.Get("/risk/overview", riskHandler.overview)
	readOnly.Get("/risk/graph/:asset_id/:threat_id", riskHandler.riskGraph)
	readOnly.Get("/risk/asset/:asset_id/attack-paths", riskHandler.assetAttackPaths)
	readOnly.Get("/risk/report/graph/:asset_id/:threat_id", riskHandler.attackPathPDF)
	readOnly.Get("/risk/report/asset/:asset_id", riskHandler.assetReportPDF)
	readOnly.Get("/threat-sources", riskHandler.listThreatSources)
	readOnly.Get("/destructive-actions", riskHandler.listDestructiveActions)

	// Compliance / ОСЗ
	readOnly.Get("/compliance/standards", complianceHandler.listStandards)
	readOnly.Get("/compliance/asset/:assetID", complianceHandler.assetOverview)
	readOnly.Get("/compliance/asset/:assetID/standard/:standardCode", complianceHandler.assetByStandard)
	readOnly.Get("/compliance/asset/:assetID/report.pdf", complianceHandler.assetReportPDF)

	// Орг.-тех. документация
	readOnly.Get("/reports/asset/:assetID/document/passport.pdf", documentHandler.passport)
	readOnly.Get("/reports/asset/:assetID/document/threat-model.pdf", documentHandler.threatModel)
	readOnly.Get("/reports/asset/:assetID/document/protection-plan.pdf", documentHandler.protectionPlan)
	readOnly.Get("/reports/asset/:assetID/documents.zip", documentHandler.pack)

	// Organization-level
	readOnly.Get("/organization/overview", organizationHandler.overview)
	readOnly.Get("/organization/asset-matrix", organizationHandler.assetMatrix)
	readOnly.Get("/organization/critical-risks", organizationHandler.criticalRisks)
	readOnly.Get("/organization/report.pdf", organizationHandler.reportPDF)

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
