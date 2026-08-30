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

// parseScale читает масштаб актива из запроса.
//
// Масштаб задаётся параметрами, а не берётся из карточки актива, сознательно:
// это позволяет проигрывать сценарии «а если станций будет двести» без правки
// данных. Незаданные значения дают по одной единице каждого вида — тогда
// стоимость совпадает со списочной ценой, как было до появления масштаба.
func parseScale(c *fiber.Ctx) optimizer.AssetScale {
	scale := optimizer.DefaultScale()
	if v, err := strconv.Atoi(c.Query("workstations")); err == nil && v > 0 {
		scale.Workstations = v
	}
	if v, err := strconv.Atoi(c.Query("servers")); err == nil && v > 0 {
		scale.Servers = v
	}
	if v, err := strconv.Atoi(c.Query("appliances")); err == nil && v > 0 {
		scale.Appliances = v
	}
	return scale
}

// optimize подбирает комплекс средств защиты для актива в рамках бюджета.
//
// Параметры:
//
//	budget       — предел затрат в рублях (обязателен);
//	max_class    — класс защиты по ФСТЭК не хуже указанного. Для ГИС высокого
//	               класса защищённости средства ниже нужного класса недопустимы,
//	               и отсекать их надо до всякой экономики;
//	workstations,
//	servers,
//	appliances   — масштаб актива: на него умножается цена за единицу.
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

	plan, err := h.svc.Optimize(c.Context(), assetID, budget, maxClass, parseScale(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(plan)
}

// roadmap строит план внедрения на несколько лет при годовом бюджете.
//
// Отличается от optimize целевой функцией: минимизируется площадь под кривой
// риска за горизонт, а не конечный риск. Из-за этого важен порядок закупок —
// средство, внедрённое в первый год, защищает дольше, чем купленное в
// последний, даже если итоговый набор совпадает.
//
// Параметры:
//
//	budget_per_year — годовой предел затрат в рублях (обязателен);
//	years           — горизонт планирования, по умолчанию 3, не больше 5;
//	max_class       — класс защиты по ФСТЭК не хуже указанного;
//	workstations,
//	servers,
//	appliances      — масштаб актива: на него умножается цена за единицу.
func (h *OptimizerHandler) roadmap(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("assetID"), 10, 64)
	if err != nil || assetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset id"})
	}

	budget, err := strconv.ParseFloat(c.Query("budget_per_year"), 64)
	if err != nil || budget <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "budget_per_year must be a positive number"})
	}

	years := 0
	if raw := c.Query("years"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "years must be a positive integer"})
		}
		years = v
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

	plan, err := h.svc.Roadmap(c.Context(), assetID, budget, years, maxClass, parseScale(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(plan)
}
