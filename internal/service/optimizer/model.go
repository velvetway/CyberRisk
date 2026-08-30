// Package optimizer подбирает комплекс средств защиты для актива.
//
// Целевая функция: максимальное снижение суммарного веса угроз W при
// ограниченном бюджете. Денежная оценка ущерба сознательно не используется —
// её пришлось бы брать экспертно, а W уже считается моделью ПТСЗИ на реальных
// данных БДУ ФСТЭК.
//
// Расчёт W оптимизатор не переопределяет: он использует ту же формулу
// ptszi.CalculateW и ту же мультипликативную модель покрытия, просто применяет
// их к гипотетическим наборам мер.
package optimizer

import "Diplom/internal/domain"

// Candidate — метод противодействия, который можно внедрить, вместе с самым
// дешёвым сертифицированным средством, которое его закрывает.
type Candidate struct {
	ControlCode string `json:"control_code"`
	ControlName string `json:"control_name"`
	// Certificate — конкретное средство из реестра ФСТЭК.
	CertificateID   int64   `json:"certificate_id"`
	ProductName     string  `json:"product_name"`
	Vendor          *string `json:"vendor,omitempty"`
	ProtectionClass *int16  `json:"protection_class,omitempty"`
	// CostMin/CostMax — диапазон, потому что цены собраны диапазоном.
	// Планирование ведём по верхней границе: бюджет должен сойтись в худшем случае.
	CostMin      float64 `json:"cost_min"`
	CostMax      float64 `json:"cost_max"`
	LicenseModel string  `json:"license_model"`
	SourceURL    *string `json:"source_url,omitempty"`
	SourceType   string  `json:"source_type"`
	// Effectiveness — предполагаемая эффективность внедрения.
	Effectiveness float64 `json:"effectiveness"`
	// ValidUntil — дата окончания сертификата ФСТЭК (ISO) либо nil у бессрочных.
	// В отличие от сроков внедрения это не экспертная оценка, а факт из реестра.
	ValidUntil *string `json:"valid_until,omitempty"`
	// ValidityKind — dated | perpetual | suspended | unknown.
	ValidityKind string `json:"validity_kind,omitempty"`
}

// Step — один шаг плана внедрения.
type Step struct {
	Candidate Candidate `json:"candidate"`
	// WBefore/WAfter — суммарный вес угроз актива до и после этого шага.
	WBefore float64 `json:"w_before"`
	WAfter  float64 `json:"w_after"`
	DeltaW  float64 `json:"delta_w"`
	// Efficiency — снижение риска на миллион рублей. Именно по этой величине
	// жадный алгоритм выбирает следующий шаг.
	Efficiency     float64 `json:"efficiency"`
	CumulativeCost float64 `json:"cumulative_cost"`
}

// Plan — результат подбора.
type Plan struct {
	AssetID int64   `json:"asset_id"`
	Budget  float64 `json:"budget"`
	// BaselineW — суммарный W по применимым угрозам до внедрения.
	BaselineW  float64 `json:"baseline_w"`
	ResultingW float64 `json:"resulting_w"`
	TotalDelta float64 `json:"total_delta"`
	TotalCost  float64 `json:"total_cost"`
	Steps      []Step  `json:"steps"`
	// Skipped — кандидаты, не вошедшие в план, и почему.
	Skipped []SkippedCandidate `json:"skipped,omitempty"`
	// Method — каким алгоритмом получен план.
	Method string `json:"method"`
	// ExhaustiveChecked — сверялся ли результат с полным перебором.
	ExhaustiveChecked bool `json:"exhaustive_checked"`
	// GreedyIsOptimal имеет смысл только вместе с ExhaustiveChecked: показывает,
	// совпал ли жадный результат с точным оптимумом.
	GreedyIsOptimal bool `json:"greedy_is_optimal"`
	// ExhaustiveDelta — снижение W у точного оптимума, для сравнения.
	ExhaustiveDelta float64 `json:"exhaustive_delta,omitempty"`
	// Warnings — оговорки, без которых план легко прочитать неверно.
	Warnings []string `json:"warnings,omitempty"`
}

type SkippedCandidate struct {
	Candidate Candidate `json:"candidate"`
	Reason    string    `json:"reason"`
}

// threatState — рабочая копия пути атаки, по которой считается W при
// гипотетическом наборе мер.
type threatState struct {
	path domain.PTSZIAttackPath
}
