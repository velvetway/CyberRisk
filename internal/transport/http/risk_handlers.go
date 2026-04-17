// internal/transport/http/risk_handlers.go
package http

import (
	"strconv"

	"Diplom/internal/report"
	"Diplom/internal/service/risk"

	"github.com/gofiber/fiber/v2"
)

type RiskHandler struct {
	svc risk.Service
}

func NewRiskHandler(svc risk.Service) *RiskHandler {
	return &RiskHandler{svc: svc}
}

func (h *RiskHandler) Register(r fiber.Router) {
	r.Post("/preview", h.previewRisk)
	r.Get("/overview", h.overview)
	r.Get("/asset/:id", h.assetRiskProfile)
	r.Post("/report/pdf", h.GenerateRiskPDF)
}

//////////////////// PREVIEW ////////////////////

type riskPreviewRequest struct {
	AssetID  int64 `json:"asset_id"`
	ThreatID int64 `json:"threat_id"`
}

type recommendationDTO struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
}

type riskPreviewResponse struct {
	AssetID         int64               `json:"asset_id"`
	ThreatID        int64               `json:"threat_id"`
	Impact          int16               `json:"impact"`
	Likelihood      int16               `json:"likelihood"`
	Score           int16               `json:"score"`
	Level           string              `json:"level"`
	Recommendations []recommendationDTO `json:"recommendations"`
}

func (h *RiskHandler) previewRisk(c *fiber.Ctx) error {
	var req riskPreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}

	if req.AssetID <= 0 || req.ThreatID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "asset_id and threat_id must be > 0"})
	}

	res, recs, err := h.svc.PreviewRisk(c.Context(), req.AssetID, req.ThreatID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	out := riskPreviewResponse{
		AssetID:    req.AssetID,
		ThreatID:   req.ThreatID,
		Impact:     res.Impact,
		Likelihood: res.Likelihood,
		Score:      res.Score,
		Level:      res.Level,
	}

	for _, r := range recs {
		out.Recommendations = append(out.Recommendations, recommendationDTO{
			Code:        r.Code,
			Title:       r.Title,
			Description: r.Description,
			Priority:    r.Priority,
			Category:    r.Category,
		})
	}

	return c.JSON(out)
}

//////////////////// OVERVIEW ////////////////////

func (h *RiskHandler) overview(c *fiber.Ctx) error {
	data, err := h.svc.Overview(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

//////////////////// ASSET RISK PROFILE ////////////////////

func (h *RiskHandler) assetRiskProfile(c *fiber.Ctx) error {
	assetID, err := c.ParamsInt("id")
	if err != nil || assetID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid asset id"})
	}

	risks, err := h.svc.AssetRiskProfile(c.Context(), int64(assetID))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(risks)
}

//////////////////// PDF-REPORT ////////////////////

type riskReportRequest struct {
	AssetID  int64 `json:"asset_id"`
	ThreatID int64 `json:"threat_id"`
}

func (h *RiskHandler) GenerateRiskPDF(c *fiber.Ctx) error {
	var req riskReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}

	if req.AssetID <= 0 || req.ThreatID <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "asset_id and threat_id required"})
	}

	// <<<<< ключевое исправление
	res, recs, err := h.svc.PreviewRisk(c.Context(), req.AssetID, req.ThreatID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	pdfBytes, err := report.GenerateRiskPDF(req.AssetID, req.ThreatID, *res, recs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=risk_report.pdf")
	return c.Send(pdfBytes)
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
