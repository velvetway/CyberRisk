// organization_handlers.go — сводка по всем активам организации.
// Закрывает потребность «не один актив = один отчёт, а вся организация»:
// агрегированные метрики, таблица всех активов с показателями, топ
// критических рисков, единый PDF-отчёт.
package http

import (
	"fmt"
	"strconv"

	"Diplom/internal/report"
	"Diplom/internal/service/organization"

	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct {
	svc organization.Service
}

func NewOrganizationHandler(svc organization.Service) *OrganizationHandler {
	return &OrganizationHandler{svc: svc}
}

// GET /api/organization/overview — сводные метрики (для дашборда).
func (h *OrganizationHandler) overview(c *fiber.Ctx) error {
	out, err := h.svc.Overview(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/organization/asset-matrix — табличный список активов с показателями.
func (h *OrganizationHandler) assetMatrix(c *fiber.Ctx) error {
	out, err := h.svc.AssetMatrix(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/organization/critical-risks?limit=N — топ-N (asset × threat) с
// наибольшим W. По умолчанию N=20.
func (h *OrganizationHandler) criticalRisks(c *fiber.Ctx) error {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	out, err := h.svc.CriticalRisks(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/organization/report.pdf — единый сводный PDF.
func (h *OrganizationHandler) reportPDF(c *fiber.Ctx) error {
	overview, err := h.svc.Overview(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	matrix, err := h.svc.AssetMatrix(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	critical, err := h.svc.CriticalRisks(c.Context(), 25)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	pdf, err := report.GenerateOrganizationReportPDF(overview, matrix, critical)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, "organization-report.pdf"))
	return c.Send(pdf)
}
