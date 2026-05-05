package http

import (
	"encoding/json"
	"strconv"

	"Diplom/internal/domain"
	assetService "Diplom/internal/service/asset"
	softwareService "Diplom/internal/service/software"

	"github.com/gofiber/fiber/v2"
)

type AssetHandler struct {
	svc         assetService.Service
	softwareSvc softwareService.Service
}

func NewAssetHandler(svc assetService.Service, softwareSvc softwareService.Service) *AssetHandler {
	return &AssetHandler{svc: svc, softwareSvc: softwareSvc}
}

func (h *AssetHandler) Register(r fiber.Router) {
	r.Get("/", h.listAssets)
	r.Post("/", h.createAsset)
	r.Get("/:id/software/alternatives", h.assetSoftwareAlternatives)
	r.Get("/:id", h.getAsset)
	r.Put("/:id", h.updateAsset)
	r.Delete("/:id", h.deleteAsset)
}

// AssetResponse is the API surface for an asset under the PTSZI W-model.
// All legacy regulatory / criticality / CIA fields have been dropped from the
// response — see docs/risk-model.md.
type AssetResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	AssetTypeID *int16                 `json:"asset_type_id,omitempty"`
	Owner       *string                `json:"owner,omitempty"`
	Description *string                `json:"description,omitempty"`
	Environment string                 `json:"environment"`
	IsIsolated  bool                   `json:"is_isolated"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

func assetToResponse(a *domain.Asset) AssetResponse {
	var tags map[string]interface{}
	if len(a.Tags) > 0 {
		_ = json.Unmarshal(a.Tags, &tags)
	}

	return AssetResponse{
		ID:          a.ID,
		Name:        a.Name,
		AssetTypeID: a.AssetTypeID,
		Owner:       a.Owner,
		Description: a.Description,
		Environment: string(a.Environment),
		IsIsolated:  a.IsIsolated,
		Tags:        tags,
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *AssetHandler) listAssets(c *fiber.Ctx) error {
	limit, _ := strconv.ParseInt(c.Query("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.Query("offset", "0"), 10, 32)

	assets, err := h.svc.List(c.Context(), int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	res := make([]AssetResponse, 0, len(assets))
	for i := range assets {
		res = append(res, assetToResponse(&assets[i]))
	}

	return c.JSON(res)
}

func (h *AssetHandler) createAsset(c *fiber.Ctx) error {
	var in assetService.CreateAssetInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	a, err := h.svc.Create(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(assetToResponse(a))
}

func (h *AssetHandler) getAsset(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	a, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if a == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "asset not found"})
	}

	return c.JSON(assetToResponse(a))
}

func (h *AssetHandler) assetSoftwareAlternatives(c *fiber.Ctx) error {
	if h.softwareSvc == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "software service not configured"})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	asset, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if asset == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "asset not found"})
	}

	alternatives, err := h.softwareSvc.SuggestAlternativesForAsset(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(alternatives)
}

func (h *AssetHandler) updateAsset(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var in assetService.UpdateAssetInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}

	a, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(assetToResponse(a))
}

func (h *AssetHandler) deleteAsset(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
