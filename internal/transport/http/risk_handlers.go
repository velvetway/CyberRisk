// internal/transport/http/risk_handlers.go
package http

import (
	"strconv"

	"Diplom/internal/service/risk"

	"github.com/gofiber/fiber/v2"
)

type RiskHandler struct {
	svc risk.Service
}

func NewRiskHandler(svc risk.Service) *RiskHandler {
	return &RiskHandler{svc: svc}
}

//////////////////// PTSZI OVERVIEW ////////////////////

func (h *RiskHandler) overview(c *fiber.Ctx) error {
	data, err := h.svc.Overview(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

//////////////////// PTSZI GRAPH ////////////////////

// GET /api/risk/graph/:asset_id/:threat_id
func (h *RiskHandler) riskGraph(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("asset_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset_id"})
	}
	threatID, err := strconv.ParseInt(c.Params("threat_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid threat_id"})
	}

	path, err := h.svc.AssembleAttackPath(c.Context(), assetID, threatID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(path)
}

// GET /api/risk/asset/:asset_id/attack-paths
func (h *RiskHandler) assetAttackPaths(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("asset_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset_id"})
	}

	res, err := h.svc.AssembleAssetAttackPaths(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// GET /api/threat-sources
func (h *RiskHandler) listThreatSources(c *fiber.Ctx) error {
	out, err := h.svc.ListThreatSources(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GET /api/destructive-actions
func (h *RiskHandler) listDestructiveActions(c *fiber.Ctx) error {
	out, err := h.svc.ListDestructiveActions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}
