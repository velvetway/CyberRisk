package http

import (
	"strconv"

	"Diplom/internal/repository"
	externalCatalog "Diplom/internal/service/external_catalog"

	"github.com/gofiber/fiber/v2"
)

type ExternalCatalogHandler struct {
	svc externalCatalog.Service
}

func NewExternalCatalogHandler(svc externalCatalog.Service) *ExternalCatalogHandler {
	return &ExternalCatalogHandler{svc: svc}
}

func (h *ExternalCatalogHandler) searchBDU(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "25"))
	var minCVSS *float64
	if raw := c.Query("min_cvss"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			minCVSS = &v
		}
	}
	var minSeverity *int16
	if raw := c.Query("min_severity"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 16); err == nil {
			x := int16(v)
			minSeverity = &x
		}
	}

	items, err := h.svc.SearchBDU(c.Context(), repository.BDUSearchFilter{
		Query:       c.Query("q"),
		Software:    c.Query("software"),
		Vendor:      c.Query("vendor"),
		MinCVSS:     minCVSS,
		MinSeverity: minSeverity,
		Limit:       limit,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *ExternalCatalogHandler) searchMinreestr(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "25"))
	var categoryID *int16
	if raw := c.Query("category_id"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 16); err == nil {
			x := int16(v)
			categoryID = &x
		}
	}

	items, err := h.svc.SearchMinreestr(c.Context(), repository.MinreestrFilter{
		Query:       c.Query("q"),
		CategoryID:  categoryID,
		RussianOnly: c.Query("russian_only") == "true",
		FSTECOnly:   c.Query("fstec_only") == "true",
		Limit:       limit,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// searchSZI — поиск по Государственному реестру сертифицированных СЗИ ФСТЭК.
// Параметр control отбирает средства по методу противодействия ПТСЗИ,
// max_class — по классу защиты (не хуже указанного, меньше значение строже).
func (h *ExternalCatalogHandler) searchSZI(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "25"))

	var maxClass *int16
	if raw := c.Query("max_class"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 16); err == nil {
			x := int16(v)
			maxClass = &x
		}
	}

	// По умолчанию отдаём только действующие сертификаты: просроченные
	// в подборе средств бесполезны.
	activeOnly := c.Query("active_only", "true") != "false"

	items, err := h.svc.SearchSZI(c.Context(), repository.SZISearchFilter{
		Query:              c.Query("q"),
		ControlCode:        c.Query("control"),
		MaxProtectionClass: maxClass,
		ActiveOnly:         activeOnly,
		Limit:              limit,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// sziCoverage — сколько сертифицированных средств доступно на каждый метод ПТСЗИ.
func (h *ExternalCatalogHandler) sziCoverage(c *fiber.Ctx) error {
	items, err := h.svc.SZIControlCoverage(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}
