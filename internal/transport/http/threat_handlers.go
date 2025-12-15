package http

import (
	"strconv"
	"time"

	"Diplom/internal/domain"
	threatService "Diplom/internal/service/threat"

	"github.com/gofiber/fiber/v2"
)

type ThreatHandler struct {
	svc threatService.Service
}

func NewThreatHandler(svc threatService.Service) *ThreatHandler {
	return &ThreatHandler{svc: svc}
}

func (h *ThreatHandler) Register(r fiber.Router) {
	r.Get("/", h.listThreats)
	r.Post("/", h.createThreat)
	r.Get("/:id", h.getThreat)
	r.Put("/:id", h.updateThreat)
	r.Delete("/:id", h.deleteThreat)
}

type ThreatResponse struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	ThreatCategoryID *int16  `json:"threat_category_id,omitempty"`
	SourceType       string  `json:"source_type"`
	Description      *string `json:"description,omitempty"`
	BaseLikelihood   int16   `json:"base_likelihood"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func threatToResponse(t *domain.Threat) ThreatResponse {
	return ThreatResponse{
		ID:               t.ID,
		Name:             t.Name,
		ThreatCategoryID: t.ThreatCategoryID,
		SourceType:       string(t.SourceType),
		Description:      t.Description,
		BaseLikelihood:   t.BaseLikelihood,
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        t.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *ThreatHandler) listThreats(c *fiber.Ctx) error {
	limit64, _ := strconv.ParseInt(c.Query("limit", "50"), 10, 32)
	offset64, _ := strconv.ParseInt(c.Query("offset", "0"), 10, 32)

	threats, err := h.svc.List(c.Context(), int32(limit64), int32(offset64))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	res := make([]ThreatResponse, 0, len(threats))
	for i := range threats {
		res = append(res, threatToResponse(&threats[i]))
	}

	return c.JSON(res)
}

func (h *ThreatHandler) createThreat(c *fiber.Ctx) error {
	var in threatService.CreateThreatInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	t, err := h.svc.Create(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(threatToResponse(t))
}

func (h *ThreatHandler) getThreat(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	t, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if t == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "threat not found"})
	}

	return c.JSON(threatToResponse(t))
}

func (h *ThreatHandler) updateThreat(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var in threatService.UpdateThreatInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	t, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(threatToResponse(t))
}

func (h *ThreatHandler) deleteThreat(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
