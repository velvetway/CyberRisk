package http

import (
	"strconv"

	ptsziService "Diplom/internal/service/ptszi"

	"github.com/gofiber/fiber/v2"
)

type PTSZIHandler struct {
	svc ptsziService.Service
}

func NewPTSZIHandler(svc ptsziService.Service) *PTSZIHandler {
	return &PTSZIHandler{svc: svc}
}

func (h *PTSZIHandler) listSources(c *fiber.Ctx) error {
	items, err := h.svc.ListSources(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) listThreats(c *fiber.Ctx) error {
	items, err := h.svc.ListThreats(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) listVulnerableLinks(c *fiber.Ctx) error {
	items, err := h.svc.ListVulnerableLinks(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) listControls(c *fiber.Ctx) error {
	items, err := h.svc.ListControls(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) listDestructiveActions(c *fiber.Ctx) error {
	items, err := h.svc.ListDestructiveActions(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) listUBI(c *fiber.Ctx) error {
	limit, _ := strconv.ParseInt(c.Query("limit", "100"), 10, 32)
	offset, _ := strconv.ParseInt(c.Query("offset", "0"), 10, 32)
	items, err := h.svc.ListUBI(c.Context(), int32(limit), int32(offset), c.Query("q"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) assetProfile(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}
	profile, err := h.svc.AssetProfile(c.Context(), assetID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(profile)
}

func (h *PTSZIHandler) applicableThreats(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}
	items, err := h.svc.ApplicableThreats(c.Context(), assetID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *PTSZIHandler) graph(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}
	threatID, err := strconv.ParseInt(c.Params("threatID"), 10, 64)
	if err != nil || threatID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid threat id"})
	}
	path, err := h.svc.AttackPath(c.Context(), assetID, threatID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(path)
}

type updateAssetVLRequest struct {
	VulnerableLinkIDs []int16 `json:"vulnerable_link_ids"`
}

func (h *PTSZIHandler) updateAssetVulnerableLinks(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}
	var req updateAssetVLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}
	if err := h.svc.UpdateAssetVulnerableLinks(c.Context(), assetID, req.VulnerableLinkIDs); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

type updateAssetControlsRequest struct {
	Controls []ptsziService.AssetControlInput `json:"controls"`
}

func (h *PTSZIHandler) updateAssetControls(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}
	var req updateAssetControlsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}
	if err := h.svc.UpdateAssetControls(c.Context(), assetID, req.Controls); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
