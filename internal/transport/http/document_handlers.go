// document_handlers.go — генерация орг.-тех. документации по активу.
// Закрывает блок «Формирование организационно-технической документации»
// из 7.png диплома.
package http

import (
	"context"
	"fmt"
	"strconv"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
	"Diplom/internal/report"
	"Diplom/internal/service/compliance"
	"Diplom/internal/service/risk"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	assetRepo     repository.AssetRepository
	softwareRepo  repository.SoftwareRepository
	controlRepo   repository.ControlRepository
	assetVulnRepo repository.AssetVulnerabilityRepository
	riskSvc       risk.Service
	complianceSvc compliance.Service
	pool          interface { // мини-репо asset_types для подстановки имени
		QueryRow(ctx context.Context, sql string, args ...any) interface {
			Scan(...any) error
		}
	}
	assetTypeName func(ctx context.Context, id int16) string
}

// NewDocumentHandler — конструктор. assetTypeName можно nil — тогда имя
// типа подставится прочерком.
func NewDocumentHandler(
	assetRepo repository.AssetRepository,
	softwareRepo repository.SoftwareRepository,
	controlRepo repository.ControlRepository,
	assetVulnRepo repository.AssetVulnerabilityRepository,
	riskSvc risk.Service,
	complianceSvc compliance.Service,
	assetTypeName func(ctx context.Context, id int16) string,
) *DocumentHandler {
	return &DocumentHandler{
		assetRepo:     assetRepo,
		softwareRepo:  softwareRepo,
		controlRepo:   controlRepo,
		assetVulnRepo: assetVulnRepo,
		riskSvc:       riskSvc,
		complianceSvc: complianceSvc,
		assetTypeName: assetTypeName,
	}
}

// ----- сборщик данных -----

type bundle struct {
	asset       *domain.Asset
	typeName    string
	software    []domain.AssetSoftwareWithSoftware
	controls    []domain.Control
	vulns       []domain.AssetVulnerability
	attackPaths *domain.AssetAttackPathsResponse
	compliance  []*domain.AssetStandardCompliance
}

func (h *DocumentHandler) collect(ctx context.Context, assetID int64) (*bundle, error) {
	asset, err := h.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset %d not found", assetID)
	}
	typeName := ""
	if asset.AssetTypeID != nil && h.assetTypeName != nil {
		typeName = h.assetTypeName(ctx, *asset.AssetTypeID)
	}

	sw, err := h.softwareRepo.ListAssetSoftware(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("software: %w", err)
	}
	ctrls, err := h.controlRepo.ListAttached(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("controls: %w", err)
	}
	vulns, err := h.assetVulnRepo.ListByAsset(ctx, assetID)
	if err != nil {
		// уязвимости опциональны — не падаем
		vulns = nil
	}
	paths, err := h.riskSvc.AssembleAssetAttackPaths(ctx, assetID)
	if err != nil {
		// риск-граф опционален — пустой ответ
		paths = nil
	}
	compl, err := h.complianceSvc.AssetAllStandards(ctx, assetID)
	if err != nil {
		compl = nil
	}

	return &bundle{
		asset:       asset,
		typeName:    typeName,
		software:    sw,
		controls:    ctrls,
		vulns:       vulns,
		attackPaths: paths,
		compliance:  compl,
	}, nil
}

func parseAssetID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid assetID")
	}
	return id, nil
}

// GET /api/reports/asset/:assetID/document/passport.pdf
func (h *DocumentHandler) passport(c *fiber.Ctx) error {
	assetID, err := parseAssetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	b, err := h.collect(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	pdf, err := report.GenerateAssetPassportPDF(&report.AssetPassportData{
		Asset:            *b.asset,
		AssetTypeName:    b.typeName,
		Software:         b.software,
		Controls:         b.controls,
		Vulnerabilities:  b.vulns,
		ComplianceScores: b.compliance,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return sendPDF(c, fmt.Sprintf("passport-asset-%d.pdf", assetID), pdf)
}

// GET /api/reports/asset/:assetID/document/threat-model.pdf
func (h *DocumentHandler) threatModel(c *fiber.Ctx) error {
	assetID, err := parseAssetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	b, err := h.collect(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	pdf, err := report.GenerateThreatModelPDF(b.asset, b.attackPaths)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return sendPDF(c, fmt.Sprintf("threat-model-asset-%d.pdf", assetID), pdf)
}

// GET /api/reports/asset/:assetID/document/protection-plan.pdf
func (h *DocumentHandler) protectionPlan(c *fiber.Ctx) error {
	assetID, err := parseAssetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	b, err := h.collect(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	paths := []domain.AttackPath{}
	if b.attackPaths != nil {
		paths = b.attackPaths.Paths
	}
	pdf, err := report.GenerateProtectionPlanPDF(&report.ProtectionPlanData{
		Asset:            *b.asset,
		Controls:         b.controls,
		AttackPaths:      paths,
		ComplianceScores: b.compliance,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return sendPDF(c, fmt.Sprintf("protection-plan-asset-%d.pdf", assetID), pdf)
}

// GET /api/reports/asset/:assetID/documents.zip — все 3 в одном архиве.
func (h *DocumentHandler) pack(c *fiber.Ctx) error {
	assetID, err := parseAssetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	b, err := h.collect(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	pa, err := report.GenerateAssetPassportPDF(&report.AssetPassportData{
		Asset:            *b.asset,
		AssetTypeName:    b.typeName,
		Software:         b.software,
		Controls:         b.controls,
		Vulnerabilities:  b.vulns,
		ComplianceScores: b.compliance,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "passport: " + err.Error()})
	}
	tm, err := report.GenerateThreatModelPDF(b.asset, b.attackPaths)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "threat-model: " + err.Error()})
	}
	paths := []domain.AttackPath{}
	if b.attackPaths != nil {
		paths = b.attackPaths.Paths
	}
	pp, err := report.GenerateProtectionPlanPDF(&report.ProtectionPlanData{
		Asset:            *b.asset,
		Controls:         b.controls,
		AttackPaths:      paths,
		ComplianceScores: b.compliance,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "protection-plan: " + err.Error()})
	}
	cp, _ := report.GenerateCompliancePDF(b.asset.Name, b.compliance)

	entries := []report.DocumentPackEntry{
		{Filename: fmt.Sprintf("01_passport.pdf"), PDF: pa},
		{Filename: fmt.Sprintf("02_threat_model.pdf"), PDF: tm},
		{Filename: fmt.Sprintf("03_protection_plan.pdf"), PDF: pp},
	}
	if cp != nil {
		entries = append(entries, report.DocumentPackEntry{Filename: "04_compliance.pdf", PDF: cp})
	}

	zipBytes, err := report.PackDocuments(entries)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="documents-asset-%d.zip"`, assetID))
	return c.Send(zipBytes)
}

func sendPDF(c *fiber.Ctx, filename string, body []byte) error {
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	return c.Send(body)
}
