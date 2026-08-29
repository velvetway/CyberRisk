package http

import (
	"strconv"

	"Diplom/internal/service/optimizer"

	"github.com/gofiber/fiber/v2"
)

type OptimizerHandler struct {
	svc optimizer.Service
}

func NewOptimizerHandler(svc optimizer.Service) *OptimizerHandler {
	return &OptimizerHandler{svc: svc}
}

// optimize подбирает комплекс средств защиты для актива в рамках бюджета.
//
// Параметры:
//
//	budget    — предел затрат в рублях (обязателен);
//	max_class — класс защиты по ФСТЭК не хуже указанного. Для ГИС высокого
//	            класса защищённости средства ниже нужного класса недопустимы,
//	            и отсекать их надо до всякой экономики.
func (h *OptimizerHandler) optimize(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset id"})
	}

	budget, err := strconv.ParseFloat(c.Query("budget"), 64)
	if err != nil || budget <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "budget must be a positive number"})
	}

	var maxClass *int16
	if raw := c.Query("max_class"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 16)
		if err != nil || v < 1 || v > 6 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "max_class must be 1..6"})
		}
		x := int16(v)
		maxClass = &x
	}

	plan, err := h.svc.Optimize(c.Context(), assetID, budget, maxClass)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(plan)
}
