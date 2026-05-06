package http

import (
	"strconv"

	"Diplom/internal/service/compliance"

	"github.com/gofiber/fiber/v2"
)

type ComplianceHandler struct {
	svc compliance.Service
}

func NewComplianceHandler(svc compliance.Service) *ComplianceHandler {
	return &ComplianceHandler{svc: svc}
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
