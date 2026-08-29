package optimizer

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
	"Diplom/internal/service/ptszi"
)

// defaultEffectiveness — эффективность вновь внедряемого средства.
//
// Модель ПТСЗИ ждёт эффективность внедрения в [0,1], но вывести её из данных
// реестра нельзя: сертификат подтверждает соответствие требованиям, а не
// качество эксплуатации. Берём осторожное значение, единое для всех средств,
// чтобы подбор сравнивал их по покрытию звеньев и цене, а не по выдуманным
// различиям в эффективности.
const defaultEffectiveness = 0.8

type Service interface {
	Optimize(ctx context.Context, assetID int64, budget float64, maxClass *int16) (*Plan, error)
}

type service struct {
	ptszi ptszi.Service
	szi   repository.SZIRepository
}

func NewService(p ptszi.Service, szi repository.SZIRepository) Service {
	return &service{ptszi: p, szi: szi}
}

func (s *service) Optimize(ctx context.Context, assetID int64, budget float64, maxClass *int16) (*Plan, error) {
	if budget <= 0 {
		return nil, fmt.Errorf("budget must be positive")
	}
	if s.szi == nil || !s.szi.IsAvailable() {
		return nil, fmt.Errorf("szi catalog is not available")
	}

	paths, err := s.ptszi.ApplicableThreats(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("applicable threats: %w", err)
	}

	plan := &Plan{
		AssetID:   assetID,
		Budget:    budget,
		BaselineW: totalW(paths, nil),
		Method:    "greedy",
		Steps:     []Step{},
	}
	if len(paths) == 0 {
		plan.ResultingW = plan.BaselineW
		return plan, nil
	}

	candidates, skipped, err := s.buildCandidates(ctx, paths, maxClass)
	if err != nil {
		return nil, err
	}
	plan.Skipped = skipped
	sortCandidates(candidates)

	steps, spent := greedy(paths, candidates, budget)
	plan.Steps = steps
	plan.TotalCost = spent
	plan.ResultingW = plan.BaselineW
	if len(steps) > 0 {
		plan.ResultingW = steps[len(steps)-1].WAfter
	}
	plan.TotalDelta = plan.BaselineW - plan.ResultingW

	// Сверяем с точным оптимумом, пока размерность позволяет: без этого
	// нельзя утверждать, что жадный план хорош, — можно лишь надеяться.
	if best, ok := exhaustive(paths, candidates, budget); ok {
		plan.ExhaustiveChecked = true
		plan.ExhaustiveDelta = best
		plan.GreedyIsOptimal = plan.TotalDelta >= best-1e-9
	}

	plan.Warnings = warnings(plan.Steps)
	return plan, nil
}

// buildCandidates собирает по одному кандидату на каждый невнедрённый метод:
// самое дешёвое сертифицированное средство, которое его закрывает.
//
// Берём самое дешёвое сознательно: задача — показать минимальную цену закрытия
// метода. Более дорогие средства того же класса дали бы ту же прибавку к
// покрытию (модель не различает продукты внутри метода), но за большие деньги.
func (s *service) buildCandidates(
	ctx context.Context,
	paths []domain.PTSZIAttackPath,
	maxClass *int16,
) ([]Candidate, []SkippedCandidate, error) {
	missing := missingControls(paths)
	candidates := make([]Candidate, 0, len(missing))
	skipped := make([]SkippedCandidate, 0)

	for _, ctrl := range missing {
		certs, err := s.szi.Search(ctx, repository.SZISearchFilter{
			ControlCode:        ctrl.Code,
			ActiveOnly:         true,
			MaxProtectionClass: maxClass,
			Limit:              200,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("search szi for %s: %w", ctrl.Code, err)
		}

		best, found := cheapest(certs, ctrl)
		if !found {
			skipped = append(skipped, SkippedCandidate{
				Candidate: Candidate{ControlCode: ctrl.Code, ControlName: ctrl.Name},
				Reason:    "нет сертифицированного средства с известной ценой",
			})
			continue
		}
		candidates = append(candidates, best)
	}

	return candidates, skipped, nil
}

// cheapest выбирает средство с минимальной верхней границей цены.
// Планируем по верхней границе: бюджет должен сойтись и в худшем случае.
func cheapest(certs []domain.SZICertificate, ctrl domain.PTSZIControl) (Candidate, bool) {
	var best Candidate
	found := false

	for _, cert := range certs {
		for _, p := range cert.Prices {
			if p.PriceMin == nil || p.PriceMax == nil {
				continue
			}
			if found && *p.PriceMax >= best.CostMax {
				continue
			}
			best = Candidate{
				ControlCode:     ctrl.Code,
				ControlName:     ctrl.Name,
				CertificateID:   cert.ID,
				ProductName:     p.ProductName,
				Vendor:          p.Vendor,
				ProtectionClass: cert.ProtectionClass,
				CostMin:         *p.PriceMin,
				CostMax:         *p.PriceMax,
				LicenseModel:    p.LicenseModel,
				SourceURL:       p.SourceURL,
				SourceType:      p.SourceType,
				Effectiveness:   defaultEffectiveness,
			}
			found = true
		}
	}

	return best, found
}

// warnings собирает оговорки к плану.
//
// Главная из них про единицы лицензирования. Цены собраны за разные единицы:
// одна за рабочее место, другая за шасси, третья за мобильное устройство.
// Оптимизатор сравнивает их напрямую и потому охотно берёт формально дешёвое,
// хотя для защиты сервера мобильная лицензия за 674 рубля не годится.
// Привести цены к общей единице нельзя, пока в модели нет масштаба актива —
// числа рабочих мест, серверов и каналов. До тех пор план читается как
// «минимальная цена закрытия метода», а не как готовая смета.
func warnings(steps []Step) []string {
	out := make([]string, 0, 2)

	models := map[string]bool{}
	for _, s := range steps {
		if s.Candidate.LicenseModel != "" {
			models[s.Candidate.LicenseModel] = true
		}
	}
	if len(models) > 1 {
		out = append(out, "в плане смешаны разные единицы лицензирования "+
			"(узел, сервер, шасси, комплект): суммарная стоимость условна, "+
			"пока в модели нет масштаба актива")
	}

	for _, s := range steps {
		if s.Candidate.SourceType == "procurement" {
			out = append(out, "часть цен взята из контрактов госзакупок: "+
				"они отражают конкретную поставку с конкретным объёмом, а не прайс")
			break
		}
	}

	return out
}

// missingControls — методы, которые встречаются в сценариях актива, но ни на
// одном звене не внедрены.
func missingControls(paths []domain.PTSZIAttackPath) []domain.PTSZIControl {
	seen := map[int16]bool{}
	implemented := map[int16]bool{}
	out := make([]domain.PTSZIControl, 0)

	for i := range paths {
		for _, vl := range paths[i].VulnerableLinks {
			for _, c := range vl.Controls {
				if c.Implemented {
					implemented[c.Control.ID] = true
					continue
				}
				if !seen[c.Control.ID] {
					seen[c.Control.ID] = true
					out = append(out, c.Control)
				}
			}
		}
	}

	filtered := out[:0]
	for _, c := range out {
		if !implemented[c.ID] {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
