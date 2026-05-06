package http

import (
	"fmt"
	"strconv"

	"Diplom/internal/repository"
	"Diplom/internal/report"
	"Diplom/internal/service/compliance"

	"github.com/gofiber/fiber/v2"
)

type ComplianceHandler struct {
	svc       compliance.Service
	assetRepo repository.AssetRepository
}

func NewComplianceHandler(svc compliance.Service, assetRepo repository.AssetRepository) *ComplianceHandler {
	return &ComplianceHandler{svc: svc, assetRepo: assetRepo}
}

// GET /api/compliance/standards — список стандартов в системе.
func (h *ComplianceHandler) listStandards(c *fiber.Ctx) error {
	out, err := h.svc.ListStandards(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/compliance/asset/:assetID — короткая сводка по всем стандартам.
func (h *ComplianceHandler) assetOverview(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	out, err := h.svc.AssetOverview(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/compliance/asset/:assetID/standard/:standardCode — детальный
// разрез по требованиям одного стандарта.
func (h *ComplianceHandler) assetByStandard(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	standardCode := c.Params("standardCode")
	if standardCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "standardCode required"})
	}
	out, err := h.svc.AssetByStandard(c.Context(), assetID, standardCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/compliance/asset/:assetID/report.pdf — сводный PDF-отчёт ОСЗ.
func (h *ComplianceHandler) assetReportPDF(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	asset, err := h.assetRepo.GetByID(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if asset == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "asset not found"})
	}
	standards, err := h.svc.AssetAllStandards(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	pdf, err := report.GenerateCompliancePDF(asset.Name, standards)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="compliance-asset-%d.pdf"`, assetID))
	return c.Send(pdf)
}
