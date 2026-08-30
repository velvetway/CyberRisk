package optimizer

import (
	"sort"
	"time"
)

// Горизонт планирования ограничен сверху: дальше пяти лет прогноз держится
// на допущениях, которые проверить нечем — сертификаты истекают, продукты
// сменяют поколения, цены пересматриваются.
const (
	maxHorizonYears     = 5
	defaultHorizonYears = 3
)

// deployMonths — срок внедрения средства по модели лицензирования.
//
// Реестр сроков внедрения не содержит, и взять их неоткуда, поэтому значения
// экспертные и намеренно грубые: программная лицензия разворачивается быстрее,
// чем поставляется и монтируется железо. Точность здесь не важна — важно, что
// купленное сегодня начинает защищать не сегодня, и план это учитывает.
var deployMonths = map[string]int{
	"per_node":   1,
	"yearly":     1,
	"perpetual":  1,
	"per_server": 2,
	"bundle":     2,
	"appliance":  3,
	"unknown":    2,
}

func deployDelay(licenseModel string) int {
	if m, ok := deployMonths[licenseModel]; ok {
		return m
	}
	return deployMonths["unknown"]
}

// Purchase — закупка одного средства в конкретном периоде.
type Purchase struct {
	Candidate Candidate `json:"candidate"`
	// ActiveFromMonth — месяц горизонта, с которого средство начинает
	// снижать риск: покупка плюс срок внедрения.
	ActiveFromMonth int     `json:"active_from_month"`
	DeployMonths    int     `json:"deploy_months"`
	Cost            float64 `json:"cost"`
	// ExpiresAtMonth — месяц, с которого средство перестаёт считаться
	// защищающим, потому что закончился сертификат ФСТЭК.
	//
	// Указатель, а не число: nil означает «сертификат бессрочный или его
	// дата неизвестна». С обычным int нулевое значение читалось бы как
	// «истёк в нулевом месяце», и забытое поле молча отключало бы средство
	// на весь горизонт.
	ExpiresAtMonth *int `json:"expires_at_month,omitempty"`
}

// monthsUntil переводит дату окончания сертификата в номер месяца горизонта.
//
// Планирование ведётся от текущего месяца, поэтому сертификат, истекающий
// через год, перестаёт действовать на 12-м месяце плана.
func monthsUntil(validUntil *string, from time.Time) *int {
	if validUntil == nil || *validUntil == "" {
		return nil
	}
	end, err := time.Parse("2006-01-02", *validUntil)
	if err != nil {
		return nil
	}
	months := int(end.Year()-from.Year())*12 + int(end.Month()) - int(from.Month())
	if months < 0 {
		months = 0
	}
	return &months
}

// Period — один год плана.
type Period struct {
	Year      int        `json:"year"`
	Purchases []Purchase `json:"purchases"`
	Spent     float64    `json:"spent"`
	// WAtStart/WAtEnd — суммарный вес угроз в начале и конце года.
	// Различаются, если купленное в этом году успело внедриться.
	WAtStart float64 `json:"w_at_start"`
	WAtEnd   float64 `json:"w_at_end"`
	// RiskArea — вклад года в площадь под кривой риска, в единицах «W · год».
	RiskArea float64 `json:"risk_area"`
}

// Roadmap — план внедрения на несколько лет.
//
// Главное отличие от статического подбора: минимизируется не конечный риск,
// а площадь под кривой риска за весь горизонт. Два плана могут прийти к
// одному остаточному риску, но тот, что снижает его раньше, оставляет
// систему под ударом меньшее время — и это должно быть видно в метрике.
type Roadmap struct {
	AssetID       int64      `json:"asset_id"`
	HorizonYears  int        `json:"horizon_years"`
	BudgetPerYear float64    `json:"budget_per_year"`
	Scale         AssetScale `json:"scale"`
	Periods       []Period   `json:"periods"`
	BaselineW     float64    `json:"baseline_w"`
	FinalW        float64    `json:"final_w"`
	TotalCost     float64    `json:"total_cost"`
	// BaselineArea — площадь, если не делать ничего: W₀ × горизонт.
	BaselineArea float64 `json:"baseline_area"`
	// RiskArea — площадь под кривой при этом плане.
	RiskArea float64 `json:"risk_area"`
	// AreaReduction — на сколько сокращена площадь. Именно эту величину
	// максимизирует планировщик.
	AreaReduction float64 `json:"area_reduction"`
	// DiscountRate — ставка приведения затрат, 0 если не применялась.
	DiscountRate float64 `json:"discount_rate"`
	// PresentValue — приведённая стоимость плана. Совпадает с TotalCost,
	// когда дисконтирование выключено.
	PresentValue float64 `json:"present_value"`
	// DegradationRate — годовая скорость старения защиты, 0 если не учтена.
	DegradationRate float64            `json:"degradation_rate"`
	Skipped         []SkippedCandidate `json:"skipped,omitempty"`
	Warnings        []string           `json:"warnings,omitempty"`
}

// monthlyW — помесячный ряд суммарного веса угроз при заданных закупках.
func monthlyW(paths pathSet, purchases []Purchase, months int, degradation float64) []float64 {
	series := make([]float64, months)
	for m := 0; m < months; m++ {
		added := map[string]float64{}
		for _, p := range purchases {
			if p.ActiveFromMonth > m {
				continue
			}
			// Сертификат кончился — средство больше не считается
			// подтверждённым, и риск возвращается.
			if p.ExpiresAtMonth != nil && m >= *p.ExpiresAtMonth {
				continue
			}
			// Защита стареет с момента, когда начала работать.
			eff := degradedEffectiveness(p.Candidate.Effectiveness, degradation, m-p.ActiveFromMonth)
			aged := p.Candidate
			aged.Effectiveness = eff
			applyCandidate(added, aged)
		}
		series[m] = paths.totalW(added)
	}
	return series
}

// riskArea — площадь под кривой риска в единицах «W · год».
func riskArea(series []float64) float64 {
	sum := 0.0
	for _, w := range series {
		sum += w / 12.0
	}
	return sum
}

// planRoadmap распределяет закупки по годам так, чтобы максимально сократить
// площадь под кривой риска.
//
// Жадный по вкладу в площадь, а не по конечному снижению W: то же средство,
// купленное в первый год, экономит больше, чем купленное в последний, и
// планировщик обязан это видеть. Поэтому кандидат оценивается по тому,
// насколько уменьшится интеграл до конца горизонта, если купить его сейчас.
func planRoadmap(paths pathSet, candidates []Candidate, budgetPerYear float64, years int, now time.Time, degradation float64) ([]Period, []Purchase) {
	months := years * 12
	purchases := make([]Purchase, 0, len(candidates))
	used := map[int]bool{}
	periods := make([]Period, 0, years)

	// Неизрасходованный остаток переходит на следующий год.
	//
	// Без переноса дорогое средство при умеренном годовом бюджете
	// недостижимо навсегда: оно не влезает ни в один год по отдельности,
	// хотя за два года деньги нашлись бы. Организации так и поступают —
	// копят на крупную закупку.
	carried := 0.0

	for year := 0; year < years; year++ {
		period := Period{Year: year + 1, Purchases: []Purchase{}}
		remaining := budgetPerYear + carried

		for {
			currentArea := riskArea(monthlyW(paths, purchases, months, degradation))

			bestIdx := -1
			bestGain := 0.0
			var bestPurchase Purchase

			for i, c := range candidates {
				if used[i] || c.TotalCost > remaining {
					continue
				}
				delay := deployDelay(c.LicenseModel)
				candidate := Purchase{
					Candidate:       c,
					DeployMonths:    delay,
					ActiveFromMonth: year*12 + delay,
					Cost:            c.TotalCost,
					ExpiresAtMonth:  monthsUntil(c.ValidUntil, now),
				}
				if candidate.ActiveFromMonth >= months {
					// Внедрится уже за горизонтом — в этом плане бесполезно.
					continue
				}
				// Сертификат истекает раньше, чем средство успеет внедриться:
				// покупать его бессмысленно, оно не отработает ни дня.
				if candidate.ExpiresAtMonth != nil && *candidate.ExpiresAtMonth <= candidate.ActiveFromMonth {
					continue
				}

				trial := append(append([]Purchase{}, purchases...), candidate)
				gain := currentArea - riskArea(monthlyW(paths, trial, months, degradation))
				if gain <= 1e-9 {
					continue
				}
				// Сокращение площади на миллион рублей.
				efficiency := gain / (c.TotalCost / 1_000_000)
				if efficiency > bestGain {
					bestIdx, bestGain, bestPurchase = i, efficiency, candidate
				}
			}

			if bestIdx < 0 {
				break
			}

			used[bestIdx] = true
			purchases = append(purchases, bestPurchase)
			period.Purchases = append(period.Purchases, bestPurchase)
			period.Spent += bestPurchase.Cost
			remaining -= bestPurchase.Cost
		}

		carried = remaining
		periods = append(periods, period)
	}

	// Заполняем помесячную динамику по годам уже готовым набором закупок.
	series := monthlyW(paths, purchases, months, degradation)
	for i := range periods {
		start := i * 12
		end := start + 12
		periods[i].WAtStart = series[start]
		periods[i].WAtEnd = series[end-1]
		periods[i].RiskArea = riskArea(series[start:end])
	}

	sort.SliceStable(purchases, func(a, b int) bool {
		return purchases[a].ActiveFromMonth < purchases[b].ActiveFromMonth
	})
	return periods, purchases
}
