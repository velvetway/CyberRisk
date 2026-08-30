package optimizer

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Roadmap(ctx context.Context, assetID int64, budgetPerYear float64, years int, maxClass *int16) (*Roadmap, error)
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

// Roadmap строит план внедрения на несколько лет.
//
// В отличие от Optimize минимизируется не конечный риск, а площадь под кривой
// риска за горизонт: год, прожитый под критической угрозой, стоит дороже года
// под средней, и порядок закупок поэтому влияет на ответ.
func (s *service) Roadmap(
	ctx context.Context,
	assetID int64,
	budgetPerYear float64,
	years int,
	maxClass *int16,
) (*Roadmap, error) {
	if budgetPerYear <= 0 {
		return nil, fmt.Errorf("budget per year must be positive")
	}
	if years <= 0 {
		years = defaultHorizonYears
	}
	if years > maxHorizonYears {
		return nil, fmt.Errorf("horizon must not exceed %d years", maxHorizonYears)
	}
	if s.szi == nil || !s.szi.IsAvailable() {
		return nil, fmt.Errorf("szi catalog is not available")
	}

	paths, err := s.ptszi.ApplicableThreats(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("applicable threats: %w", err)
	}

	baseline := totalW(paths, nil)
	roadmap := &Roadmap{
		AssetID:       assetID,
		HorizonYears:  years,
		BudgetPerYear: budgetPerYear,
		BaselineW:     baseline,
		FinalW:        baseline,
		BaselineArea:  baseline * float64(years),
		RiskArea:      baseline * float64(years),
		Periods:       []Period{},
	}
	if len(paths) == 0 {
		return roadmap, nil
	}

	candidates, skipped, err := s.buildCandidates(ctx, paths, maxClass)
	if err != nil {
		return nil, err
	}
	roadmap.Skipped = skipped
	sortCandidates(candidates)

	periods, purchases := planRoadmap(paths, candidates, budgetPerYear, years, time.Now())
	roadmap.Periods = periods

	for _, p := range periods {
		roadmap.TotalCost += p.Spent
	}

	// Площадь считаем по помесячному ряду целиком, а не сложением по годам:
	// так не накапливается ошибка округления.
	series := monthlyW(paths, purchases, years*12)
	roadmap.RiskArea = riskArea(series)
	roadmap.AreaReduction = roadmap.BaselineArea - roadmap.RiskArea
	if len(series) > 0 {
		roadmap.FinalW = series[len(series)-1]
	}

	steps := make([]Step, 0, len(purchases))
	for _, p := range purchases {
		steps = append(steps, Step{Candidate: p.Candidate})
	}
	roadmap.Warnings = append(warnings(steps), expiryWarnings(purchases, years*12)...)

	return roadmap, nil
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
				ValidUntil:      cert.ValidUntil,
				ValidityKind:    cert.ValidityKind,
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

// expiryWarnings предупреждает о средствах, сертификат которых истекает
// внутри горизонта планирования.
//
// Это не гипотеза, а дата из реестра: после неё средство перестаёт быть
// подтверждённым, и защита, на которую рассчитывает план, юридически
// заканчивается. Продление сертификата — отдельное событие, которого модель
// не предсказывает, поэтому решение остаётся за человеком.
func expiryWarnings(purchases []Purchase, horizonMonths int) []string {
	expiring := make([]string, 0)
	for _, p := range purchases {
		if p.ExpiresAtMonth == nil || *p.ExpiresAtMonth >= horizonMonths {
			continue
		}
		name := p.Candidate.ProductName
		if name == "" {
			name = p.Candidate.ControlCode
		}
		until := ""
		if p.Candidate.ValidUntil != nil {
			until = " (до " + *p.Candidate.ValidUntil + ")"
		}
		expiring = append(expiring, name+until)
	}
	if len(expiring) == 0 {
		return nil
	}
	return []string{
		"внутри горизонта истекает сертификат: " + strings.Join(expiring, "; ") +
			" — после этой даты средство перестаёт быть подтверждённым, " +
			"план учитывает возврат риска",
	}
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
