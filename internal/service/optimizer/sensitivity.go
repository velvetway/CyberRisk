package optimizer

import (
	"math"
	"math/rand"
	"sort"

	"Diplom/internal/domain"
)

// Коэффициенты модели — покрытие звена мерой, эффективность внедрения,
// вероятность и опасность угрозы — заданы экспертно. Отсюда законный вопрос:
// не развалится ли план, если сдвинуть их на разумную величину?
//
// Анализ чувствительности отвечает на него численно: коэффициенты много раз
// случайно возмущаются в заданном коридоре, план пересчитывается, и видно,
// насколько устойчив его состав и результат.

const (
	defaultSensitivityRuns      = 200
	defaultSensitivityVariation = 0.2
	maxSensitivityRuns          = 2000
)

// ControlStability — как часто метод попадал в план при возмущённых
// коэффициентах.
type ControlStability struct {
	ControlCode string `json:"control_code"`
	ProductName string `json:"product_name,omitempty"`
	// Frequency — доля прогонов, где метод вошёл в план, от 0 до 1.
	// Единица означает, что решение не зависит от точности коэффициентов.
	Frequency float64 `json:"frequency"`
	Runs      int     `json:"runs"`
}

// SensitivityReport — устойчивость плана к неточности коэффициентов.
type SensitivityReport struct {
	AssetID int64   `json:"asset_id"`
	Budget  float64 `json:"budget"`
	Runs    int     `json:"runs"`
	// Variation — коридор возмущения, доля: 0.2 означает ±20%.
	Variation float64 `json:"variation"`

	// BaseDelta — снижение риска на исходных коэффициентах.
	BaseDelta float64 `json:"base_delta"`
	MeanDelta float64 `json:"mean_delta"`
	MinDelta  float64 `json:"min_delta"`
	MaxDelta  float64 `json:"max_delta"`
	// StdDev — разброс результата. Чем меньше, тем спокойнее можно
	// опираться на конкретную цифру снижения риска.
	StdDev float64 `json:"std_dev"`

	// CompositionStability — доля прогонов, где состав плана совпал с
	// исходным. Это главная величина отчёта: она отвечает не «насколько
	// точна цифра», а «то же ли самое надо покупать».
	CompositionStability float64 `json:"composition_stability"`

	// Controls — устойчивость каждого метода по отдельности. Методы с
	// частотой около единицы можно закупать не сомневаясь; те, что
	// появляются в половине прогонов, — предмет отдельного решения.
	Controls []ControlStability `json:"controls"`

	Verdict string `json:"verdict"`
}

// perturb возвращает копию сценариев со случайно сдвинутыми коэффициентами.
//
// Возмущаются те величины, которые в модели заданы экспертно: покрытие звена
// мерой, эффективность уже внедрённых средств, вероятность и опасность угрозы.
// Z не трогаем — он не оценка, а правило: единица, если угроза актуальна для
// обоих контуров, иначе половина.
func perturb(paths pathSet, variation float64, rng *rand.Rand) pathSet {
	factor := func() float64 {
		return 1 + (rng.Float64()*2-1)*variation
	}

	out := make(pathSet, len(paths))
	for i, p := range paths {
		clone := p
		clone.QThreat = clamp01(p.QThreat * factor())
		clone.QSeverity = clamp01(p.QSeverity * factor())

		clone.VulnerableLinks = make([]domain.PTSZIPathVL, len(p.VulnerableLinks))
		for j, vl := range p.VulnerableLinks {
			vlClone := vl
			vlClone.Controls = make([]domain.PTSZIControlCoverage, len(vl.Controls))
			for k, c := range vl.Controls {
				cClone := c
				cClone.Coverage = clamp01(c.Coverage * factor())
				if c.Implemented {
					cClone.Effectiveness = clamp01(c.Effectiveness * factor())
				}
				vlClone.Controls[k] = cClone
			}
			clone.VulnerableLinks[j] = vlClone
		}
		out[i] = clone
	}
	return out
}

// planComposition — состав плана как отсортированный список методов.
// Сравниваем именно состав, а не порядок: для закупки важно, что купить,
// а не в каком порядке жадный алгоритм это перебрал.
func planComposition(steps []Step) []string {
	codes := make([]string, 0, len(steps))
	for _, s := range steps {
		codes = append(codes, s.Candidate.ControlCode)
	}
	sort.Strings(codes)
	return codes
}

func sameComposition(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// analyzeSensitivity прогоняет подбор на возмущённых коэффициентах.
//
// Генератор случайных чисел детерминирован: один и тот же запрос даёт один и
// тот же отчёт. Иначе устойчивость нельзя было бы обсуждать — цифры менялись
// бы при каждом обновлении страницы.
func analyzeSensitivity(
	paths pathSet,
	candidates []Candidate,
	budget float64,
	runs int,
	variation float64,
	seed int64,
) SensitivityReport {
	if runs <= 0 {
		runs = defaultSensitivityRuns
	}
	if runs > maxSensitivityRuns {
		runs = maxSensitivityRuns
	}
	if variation <= 0 || variation > 1 {
		variation = defaultSensitivityVariation
	}

	baseSteps, _ := greedy(paths, candidates, budget)
	baseComposition := planComposition(baseSteps)
	baseDelta := 0.0
	if len(baseSteps) > 0 {
		baseDelta = paths.totalW(nil) - baseSteps[len(baseSteps)-1].WAfter
	}

	report := SensitivityReport{
		Runs:      runs,
		Variation: variation,
		BaseDelta: baseDelta,
		MinDelta:  math.Inf(1),
		MaxDelta:  math.Inf(-1),
	}

	rng := rand.New(rand.NewSource(seed))
	deltas := make([]float64, 0, runs)
	frequency := map[string]int{}
	products := map[string]string{}
	stable := 0

	for i := 0; i < runs; i++ {
		p := perturb(paths, variation, rng)
		steps, _ := greedy(p, candidates, budget)

		delta := 0.0
		if len(steps) > 0 {
			delta = p.totalW(nil) - steps[len(steps)-1].WAfter
		}
		deltas = append(deltas, delta)
		report.MinDelta = math.Min(report.MinDelta, delta)
		report.MaxDelta = math.Max(report.MaxDelta, delta)

		for _, s := range steps {
			frequency[s.Candidate.ControlCode]++
			products[s.Candidate.ControlCode] = s.Candidate.ProductName
		}
		if sameComposition(planComposition(steps), baseComposition) {
			stable++
		}
	}

	sum := 0.0
	for _, d := range deltas {
		sum += d
	}
	report.MeanDelta = sum / float64(len(deltas))

	variance := 0.0
	for _, d := range deltas {
		variance += (d - report.MeanDelta) * (d - report.MeanDelta)
	}
	report.StdDev = math.Sqrt(variance / float64(len(deltas)))
	report.CompositionStability = float64(stable) / float64(runs)

	for code, n := range frequency {
		report.Controls = append(report.Controls, ControlStability{
			ControlCode: code,
			ProductName: products[code],
			Frequency:   float64(n) / float64(runs),
			Runs:        n,
		})
	}
	sort.Slice(report.Controls, func(i, j int) bool {
		if report.Controls[i].Frequency != report.Controls[j].Frequency {
			return report.Controls[i].Frequency > report.Controls[j].Frequency
		}
		return report.Controls[i].ControlCode < report.Controls[j].ControlCode
	})

	report.Verdict = verdict(report.CompositionStability)
	if math.IsInf(report.MinDelta, 1) {
		report.MinDelta = 0
	}
	if math.IsInf(report.MaxDelta, -1) {
		report.MaxDelta = 0
	}
	return report
}

func verdict(stability float64) string {
	switch {
	case stability >= 0.9:
		return "план устойчив: состав закупки почти не зависит от точности коэффициентов"
	case stability >= 0.7:
		return "план в основном устойчив: состав меняется в меньшинстве прогонов"
	case stability >= 0.4:
		return "план чувствителен к коэффициентам: состав закупки заметно плавает, " +
			"стоит уточнить оценки покрытия и эффективности"
	default:
		return "план неустойчив: при таком разбросе коэффициентов выбор средств " +
			"определяется не данными, а случайностью оценок"
	}
}
