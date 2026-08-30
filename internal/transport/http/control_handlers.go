package http

import (
	"strconv"

	"Diplom/internal/domain"
	controlService "Diplom/internal/service/control"

	"github.com/gofiber/fiber/v2"
)

type ControlHandler struct {
	svc controlService.Service
}

func NewControlHandler(svc controlService.Service) *ControlHandler {
	return &ControlHandler{svc: svc}
}

// list — GET /api/controls
func (h *ControlHandler) list(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if items == nil {
		items = []domain.Control{}
	}
	return c.JSON(items)
}

// listForAsset — GET /api/assets/:assetID/controls
func (h *ControlHandler) listForAsset(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	items, err := h.svc.ListForAsset(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if items == nil {
		items = []domain.Control{}
	}
	return c.JSON(items)
}

type attachControlRequest struct {
	ControlID int64 `json:"control_id"`
}

// attachToAsset — POST /api/assets/:assetID/controls
func (h *ControlHandler) attachToAsset(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	var req attachControlRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	if req.ControlID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "control_id is required"})
	}
	if err := h.svc.Attach(c.Context(), assetID, req.ControlID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusCreated)
}

// detachFromAsset — DELETE /api/assets/:assetID/controls/:controlID
func (h *ControlHandler) detachFromAsset(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assetID"})
	}
	controlID, err := strconv.ParseInt(c.Params("controlID"), 10, 64)
	if err != nil || controlID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid controlID"})
	}
	if err := h.svc.Detach(c.Context(), assetID, controlID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
