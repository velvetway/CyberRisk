package main

import (
	"regexp"
	"strings"
)

// =====================================================================
// Source mapping: «Источник угрозы» free text → S1..S4 codes + source_type
// =====================================================================

type sourceClassification struct {
	SourceType   string   // "external" | "internal" | "third_party"
	SourceCodes  []string // S1..S4
	ThreatPower  string   // "low" | "medium" | "high"
}

func classifySource(raw string) sourceClassification {
	low := strings.ToLower(raw)
	hasExternal := strings.Contains(low, "внешний нарушитель")
	hasInternal := strings.Contains(low, "внутренний нарушитель")

	c := sourceClassification{ThreatPower: highestPower(low)}

	switch {
	case hasExternal && hasInternal:
		c.SourceType = "external" // mixed; we mark as external because it's broader
		c.SourceCodes = []string{"S1", "S2", "S3", "S4"}
	case hasExternal:
		c.SourceType = "external"
		c.SourceCodes = []string{"S1", "S4"} // конкуренты + хакеры
	case hasInternal:
		c.SourceType = "internal"
		c.SourceCodes = []string{"S2", "S3"} // партнёры + персонал
	default:
		// фразы «оператор связи», «третья сторона» и т.п.
		c.SourceType = "third_party"
		c.SourceCodes = []string{"S1", "S2"}
	}
	return c
}

func highestPower(low string) string {
	// Higher takes precedence — many threats list multiple intruders.
	switch {
	case strings.Contains(low, "высоким потенциалом"):
		return "high"
	case strings.Contains(low, "средним потенциалом"):
		return "medium"
	case strings.Contains(low, "низким потенциалом"):
		return "low"
	default:
		return "medium"
	}
}

// =====================================================================
// Asset type mapping: «Объект воздействия» → list of asset_types.name
//
// We match against the canonical English asset_types names from the
// 002_seed_data migration: Server, Database, Application, Network,
// Workstation, Mobile, IoT, Cloud.
// =====================================================================

type assetTypeRule struct {
	pattern *regexp.Regexp
	types   []string
}

var assetTypeRules = []assetTypeRule{
	{regexp.MustCompile(`(?i)\bгрид[\s-]?систем|облачн|cloud`), []string{"Cloud", "Server"}},
	{regexp.MustCompile(`(?i)BIOS|UEFI|микропрограмм`), []string{"Server", "Workstation"}},
	{regexp.MustCompile(`(?i)СУБД|database|базы\s*данн|реляционн`), []string{"Database"}},
	{regexp.MustCompile(`(?i)веб[\s-]*(сервер|приложен)|web[\s-]*(server|application)|HTTP[\s-]*сервер`), []string{"Server", "Application"}},
	{regexp.MustCompile(`(?i)сетев(ой|ого)\s*(трафик|пакет|оборудован)|маршрут|шлюз|коммутатор|VPN|МСЭ|firewall`), []string{"Network"}},
	{regexp.MustCompile(`(?i)АРМ\b|рабоч.{0,5}\s*станц|workstation|пользовательск.{0,5}\s*ПК|узел`), []string{"Workstation"}},
	{regexp.MustCompile(`(?i)мобильн|смартфон|планшет|телефон`), []string{"Mobile"}},
	{regexp.MustCompile(`(?i)IoT|интернет[\s-]*вещ|умн.{0,3}\s*устрой|АСУ\s*ТП`), []string{"IoT"}},
	{regexp.MustCompile(`(?i)гипервизор|виртуальн.{0,3}\s*машин|контейнер`), []string{"Server", "Cloud"}},
	{regexp.MustCompile(`(?i)систем(а|ы|ное)\s*(программн|ПО)|операционн.{0,3}\s*систем|ОС\b`), []string{"Server", "Workstation"}},
	{regexp.MustCompile(`(?i)клиентск.{0,5}\s*приложен|интерфейс\s*пользоват`), []string{"Application", "Workstation"}},
	{regexp.MustCompile(`(?i)приложен(ие|ия)|application`), []string{"Application"}},
	{regexp.MustCompile(`(?i)\bсервер|server|вычислит.{0,3}\s*ресурс`), []string{"Server"}},
}

// deriveAssetTypeNames returns the deduplicated asset type names matched in
// the target text. Empty result == "no constraints" (applies everywhere).
func deriveAssetTypeNames(target string) []string {
	if target == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, r := range assetTypeRules {
		if r.pattern.MatchString(target) {
			for _, name := range r.types {
				if !seen[name] {
					seen[name] = true
					out = append(out, name)
				}
			}
		}
	}
	return out
}

// =====================================================================
// Threat category mapping: threat name → existing threat_categories.name
// (catalogue populated by 002_seed_data + 005_bdu_fstec_threats)
// =====================================================================

type categoryRule struct {
	pattern *regexp.Regexp
	name    string
}

var categoryRules = []categoryRule{
	{regexp.MustCompile(`(?i)вирус|вредонос|malware|трояе?н|шпион|червь`), "Вредоносное ПО"},
	{regexp.MustCompile(`(?i)DDoS|DoS|отказ\s*в\s*обслуж|переполнен|исчерпан\s*ресурс`), "Отказ в обслуживании"},
	{regexp.MustCompile(`(?i)НСД|несанкциониров.{0,4}\s*доступ|подбор\s*паролей|подбор\s*учётн|повышен.{0,3}\s*привилег`), "Несанкционированный доступ"},
	{regexp.MustCompile(`(?i)фишинг|phishing|социальн.{0,5}\s*инжен`), "Социальная инженерия"},
	{regexp.MustCompile(`(?i)физическ.{0,5}\s*(доступ|воздейств)|кража\s*оборудов|съём\s*аккумулятор`), "Физический доступ"},
	{regexp.MustCompile(`(?i)утечк|перехват|съём\s*информац|раскрыт.{0,3}\s*информац|разглашен`), "Утечка информации"},
	{regexp.MustCompile(`(?i)модификац|подмен|искажен|нарушен.{0,3}\s*целостн`), "Нарушение целостности"},
	{regexp.MustCompile(`(?i)сетев.{0,3}\s*атак|сканирован|перебор\s*портов|man[\s-]*in[\s-]*the[\s-]*middle|MITM`), "Сетевые атаки"},
}

// matchCategoryName returns the first category whose pattern matches threat
// name or description, or "" if nothing matched.
func matchCategoryName(threatName, threatDesc string) string {
	combined := threatName + " " + threatDesc
	for _, r := range categoryRules {
		if r.pattern.MatchString(combined) {
			return r.name
		}
	}
	return ""
}

// =====================================================================
// Q-параметры derivation
// =====================================================================

func deriveQThreat(power string) float64 {
	switch power {
	case "low":
		return 0.3
	case "high":
		return 0.9
	default: // medium / unknown
		return 0.6
	}
}

func deriveQSeverity(impactC, impactI, impactA bool) float64 {
	cnt := 0
	for _, b := range []bool{impactC, impactI, impactA} {
		if b {
			cnt++
		}
	}
	switch cnt {
	case 3:
		return 0.9
	case 2:
		return 0.7
	case 1:
		return 0.5
	default:
		return 0.3
	}
}
