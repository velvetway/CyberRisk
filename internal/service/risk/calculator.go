package risk

import (
	"Diplom/internal/domain"
	"math"
	"strings"
)

type RiskResult struct {
	Impact           int16   `json:"impact"`
	Likelihood       int16   `json:"likelihood"`
	Score            int16   `json:"score"`
	Level            string  `json:"level"`
	RegulatoryFactor float64 `json:"regulatory_factor"` // Множитель для регуляторных требований
	AdjustedScore    float64 `json:"adjusted_score"`    // Скорректированная оценка
}

// Calculator инкапсулирует параметры формул на случай,
// если потом захотим их сделать настраиваемыми.
type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate считает риск для связки "актив + угроза + набор уязвимостей".
// Формула: Risk = Impact × Likelihood × RegulatoryFactor
func (c *Calculator) Calculate(asset *domain.Asset, threat *domain.Threat, vulns []domain.Vulnerability) RiskResult {
	maxSeverity := maxVulnSeverity(vulns)

	impact := calculateImpact(asset, threat, maxSeverity)
	likelihood := calculateLikelihood(asset, threat, maxSeverity)

	score := impact * likelihood
	level := riskLevel(score)

	// Вычисляем регуляторный множитель
	regFactor := calculateRegulatoryFactor(asset)
	adjustedScore := float64(score) * regFactor

	return RiskResult{
		Impact:           impact,
		Likelihood:       likelihood,
		Score:            score,
		Level:            level,
		RegulatoryFactor: regFactor,
		AdjustedScore:    adjustedScore,
	}
}

func maxVulnSeverity(vulns []domain.Vulnerability) int16 {
	var max int16 = 1
	for i := range vulns {
		if vulns[i].Severity > max {
			max = vulns[i].Severity
		}
	}
	return max
}

func calculateImpact(asset *domain.Asset, threat *domain.Threat, maxSeverity int16) int16 {
	// cia_avg = (C + I + A) / 3
	ciaAvg := float64(asset.Confidentiality+asset.Integrity+asset.Availability) / 3.0

	impactBase := 0.6*float64(asset.BusinessCriticality) + 0.4*ciaAvg
	severityDelta := 0.3 * float64(maxSeverity-3) // от -0.6 до +0.6

	// Учитываем влияние угрозы на CIA
	if threat != nil {
		ciaImpactBonus := 0.0
		if threat.ImpactConfidentiality && asset.Confidentiality >= 4 {
			ciaImpactBonus += 0.3
		}
		if threat.ImpactIntegrity && asset.Integrity >= 4 {
			ciaImpactBonus += 0.3
		}
		if threat.ImpactAvailability && asset.Availability >= 4 {
			ciaImpactBonus += 0.3
		}
		impactBase += ciaImpactBonus
	}

	impactRaw := impactBase + severityDelta

	return clampTo1_5(roundToInt16(impactRaw))
}

func calculateLikelihood(asset *domain.Asset, threat *domain.Threat, maxSeverity int16) int16 {
	likelihoodBase := float64(threat.BaseLikelihood)

	severityDelta := 0.2 * float64(maxSeverity-3) // от -0.4 до +0.4
	envDelta := envFactor(string(asset.Environment))

	// Учитываем изоляцию сети
	if asset.IsIsolated {
		likelihoodBase -= 1.0 // Изолированные системы менее подвержены сетевым атакам
	}

	// Учитываем наличие интернета
	if !asset.HasInternetAccess && threat.AttackVector != nil {
		if strings.ToLower(*threat.AttackVector) == "сетевой" {
			likelihoodBase -= 0.5 // Сетевые атаки менее вероятны без интернета
		}
	}

	likelihoodRaw := likelihoodBase + severityDelta + envDelta

	return clampTo1_5(roundToInt16(likelihoodRaw))
}

// calculateRegulatoryFactor вычисляет множитель риска на основе регуляторных требований
func calculateRegulatoryFactor(asset *domain.Asset) float64 {
	factor := 1.0

	// КИИ (187-ФЗ) — самые высокие требования
	if asset.KIICategory != nil {
		switch *asset.KIICategory {
		case domain.KIICategoryCat1:
			factor = math.Max(factor, 2.0) // Категория 1 — максимальная значимость
		case domain.KIICategoryCat2:
			factor = math.Max(factor, 1.7)
		case domain.KIICategoryCat3:
			factor = math.Max(factor, 1.4)
		}
	}

	// Уровень защищённости ПДн (152-ФЗ, ПП-1119)
	if asset.ProtectionLevel != nil {
		switch *asset.ProtectionLevel {
		case domain.ProtectionLevelUZ1:
			factor = math.Max(factor, 1.8) // УЗ-1 — максимальные требования
		case domain.ProtectionLevelUZ2:
			factor = math.Max(factor, 1.5)
		case domain.ProtectionLevelUZ3:
			factor = math.Max(factor, 1.3)
		case domain.ProtectionLevelUZ4:
			factor = math.Max(factor, 1.1)
		}
	}

	// Категория данных
	if asset.DataCategory != nil {
		switch *asset.DataCategory {
		case domain.DataCategoryStateSecret:
			factor = math.Max(factor, 2.5) // Гостайна — максимальный множитель
		case domain.DataCategoryPersonalDataSpec, domain.DataCategoryPersonalDataBio:
			factor = math.Max(factor, 1.6) // Специальные/биометрические ПДн
		case domain.DataCategoryBankingSecret, domain.DataCategoryMedicalSecret:
			factor = math.Max(factor, 1.5) // Банковская/врачебная тайна
		case domain.DataCategoryPersonalData:
			factor = math.Max(factor, 1.3) // Обычные ПДн
		case domain.DataCategoryCommercialSecret:
			factor = math.Max(factor, 1.2) // Коммерческая тайна
		case domain.DataCategoryConfidential:
			factor = math.Max(factor, 1.1)
		}
	}

	// Дополнительный множитель для ПДн по объёму
	if asset.HasPersonalData && asset.PersonalDataVolume != nil {
		switch *asset.PersonalDataVolume {
		case ">100000":
			factor *= 1.2 // Более 100 000 субъектов
		case "1000-100000":
			factor *= 1.1 // От 1000 до 100 000
		}
	}

	return factor
}

func envFactor(env string) float64 {
	switch strings.ToLower(env) {
	case "prod", "production":
		return 0.5
	case "stage", "staging", "preprod":
		return 0.2
	case "test":
		return -0.5
	case "dev", "development":
		return 0.0
	default:
		return 0.0
	}
}

func riskLevel(score int16) string {
	switch {
	case score >= 16:
		return "Critical"
	case score >= 11:
		return "High"
	case score >= 6:
		return "Medium"
	default:
		return "Low"
	}
}

// AdjustedRiskLevel возвращает уровень риска с учётом регуляторного множителя
func AdjustedRiskLevel(adjustedScore float64) string {
	switch {
	case adjustedScore >= 32: // 16 * 2.0
		return "Critical"
	case adjustedScore >= 22: // 11 * 2.0
		return "High"
	case adjustedScore >= 12: // 6 * 2.0
		return "Medium"
	default:
		return "Low"
	}
}

func roundToInt16(v float64) int16 {
	return int16(math.Round(v))
}

func clampTo1_5(v int16) int16 {
	if v < 1 {
		return 1
	}
	if v > 5 {
		return 5
	}
	return v
}
