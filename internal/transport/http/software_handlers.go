package http

import (
	"strconv"
	"strings"

	softwareService "Diplom/internal/service/software"

	"github.com/gofiber/fiber/v2"
)

type SoftwareHandler struct {
	svc softwareService.Service
}

func NewSoftwareHandler(svc softwareService.Service) *SoftwareHandler {
	return &SoftwareHandler{svc: svc}
}

func (h *SoftwareHandler) Register(r fiber.Router) {
	r.Get("/", h.listSoftware)
	r.Post("/", h.createSoftware)
	r.Get("/categories", h.listCategories)
	r.Get("/russian", h.listRussianSoftware)
	r.Get("/certified", h.listCertifiedSoftware)
	r.Get("/asset/:assetID/alternatives", h.assetAlternatives)
	r.Get("/:id", h.getSoftware)
	r.Get("/:id/alternatives", h.getAlternatives)
	r.Put("/:id", h.updateSoftware)
	r.Delete("/:id", h.deleteSoftware)
}

func (h *SoftwareHandler) listSoftware(c *fiber.Ctx) error {
	limit, _ := strconv.ParseInt(c.Query("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.Query("offset", "0"), 10, 32)

	var categoryID *int16
	if cat := c.Query("category_id"); cat != "" {
		if v, err := strconv.ParseInt(cat, 10, 16); err == nil {
			catVal := int16(v)
			categoryID = &catVal
		}
	}

	var isRussian *bool
	if rus := c.Query("is_russian"); rus != "" {
		val := rus == "true" || rus == "1"
		isRussian = &val
	}

	fstecOnly := c.Query("fstec_only") == "true"
	search := c.Query("search")

	software, err := h.svc.List(c.Context(), softwareService.ListFilter{
		Limit:       int32(limit),
		Offset:      int32(offset),
		CategoryID:  categoryID,
		IsRussian:   isRussian,
		FSTECOnly:   fstecOnly,
		SearchQuery: search,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(software)
}

func (h *SoftwareHandler) createSoftware(c *fiber.Ctx) error {
	var in softwareService.CreateSoftwareInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	sw, err := h.svc.Create(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(sw)
}

func (h *SoftwareHandler) getSoftware(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	sw, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if sw == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "software not found"})
	}

	return c.JSON(sw)
}

func (h *SoftwareHandler) updateSoftware(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var in softwareService.UpdateSoftwareInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	sw, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(sw)
}

func (h *SoftwareHandler) deleteSoftware(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SoftwareHandler) listCategories(c *fiber.Ctx) error {
	categories, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

func (h *SoftwareHandler) listRussianSoftware(c *fiber.Ctx) error {
	var categoryID int16 = 0
	if cat := c.Query("category_id"); cat != "" {
		if v, err := strconv.ParseInt(cat, 10, 16); err == nil {
			categoryID = int16(v)
		}
	}

	software, err := h.svc.FindRussianAlternatives(c.Context(), categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(software)
}

func (h *SoftwareHandler) listCertifiedSoftware(c *fiber.Ctx) error {
	var categoryID int16 = 0
	if cat := c.Query("category_id"); cat != "" {
		if v, err := strconv.ParseInt(cat, 10, 16); err == nil {
			categoryID = int16(v)
		}
	}

	fstecOnly := c.Query("fstec_only") == "true"

	software, err := h.svc.FindCertifiedSoftware(c.Context(), categoryID, fstecOnly)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(software)
}

func (h *SoftwareHandler) getAlternatives(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	alternatives, err := h.svc.SuggestAlternatives(c.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "software not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "software not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(alternatives)
}

func (h *SoftwareHandler) assetAlternatives(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset id"})
	}

	alternatives, err := h.svc.SuggestAlternativesForAsset(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(alternatives)
}
