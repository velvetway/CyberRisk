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
// VL-категории derivation: threat name + description + target → set of VL codes
//
// VL1..VL6 codes из vl_categories (см. migrations/034_vl_categories.up.sql).
// Возвращаем коды (а не id), потому что id зависит от порядка вставки в БД,
// а коды стабильны.
//
// Если ни одно правило не сработало, дефолт — VL2 («устаревшее ПО / версии
// с уязвимостями»). Это самая общая категория: любая угроза, эксплуатирующая
// конкретный CVE, фактически проходит через VL2.
// =====================================================================

type vlRule struct {
	pattern *regexp.Regexp
	codes   []string
}

var vlRules = []vlRule{
	// VL1 — нештатное доп. ПО (драйверы, утилиты, неавторизованные скрипты)
	{regexp.MustCompile(`(?i)драйвер|утилит|скрипт|неавторизован|неподписан|инсталляц|устанавл`), []string{"VL1"}},
	// VL2 — устаревшие версии ПО / эксплуатация уязвимостей
	{regexp.MustCompile(`(?i)устарев|outdated|необновлен|уязвим|exploit|patch|CVE|БДУ`), []string{"VL2"}},
	// VL3 — недекларируемое ПО / закладки / backdoor
	{regexp.MustCompile(`(?i)недекларир|backdoor|закладк|trojan|шпион|spyware|программ.{0,5}\s*заклад`), []string{"VL3"}},
	// VL4 — обход админом, повышение привилегий, misconfiguration
	{regexp.MustCompile(`(?i)повышен.{0,3}\s*привилег|обход\s*(полит|правил)|misconfig|администратор.{0,5}\s*(обход|нарушен)|превышен.{0,3}\s*полномоч`), []string{"VL4"}},
	// VL5 — съёмные носители
	{regexp.MustCompile(`(?i)съём.{0,5}\s*носител|флеш|USB|removable|накопител|внешн.{0,3}\s*носител`), []string{"VL5"}},
	// VL6 — открытые ОС / отсутствие защиты ЛВС / сетевые атаки
	{regexp.MustCompile(`(?i)сетев.{0,3}\s*(атак|трафик|сегмент)|перехват\s*(пакет|данн|информац)|сканирован|открыт.{0,5}\s*порт|untrusted\s*network|отсутстви.{0,3}\s*(МСЭ|firewall|сегмент)|DDoS`), []string{"VL6"}},
	// Доп. ловушки: вредонос/вирус — обычно проходит через VL1 + VL2
	{regexp.MustCompile(`(?i)вирус|вредонос|malware|червь`), []string{"VL1", "VL2"}},
	// Социальная инженерия — попадает на VL4 (обход политик пользователем)
	{regexp.MustCompile(`(?i)фишинг|phishing|социальн.{0,5}\s*инжен`), []string{"VL4"}},
}

// deriveVLCategoryCodes возвращает дедуплицированный набор VL-кодов,
// сработавших на тексте угрозы. Пустой результат → дефолт VL2.
func deriveVLCategoryCodes(threatName, threatDesc, target string) []string {
	combined := threatName + "\n" + threatDesc + "\n" + target
	seen := map[string]bool{}
	var out []string
	for _, r := range vlRules {
		if r.pattern.MatchString(combined) {
			for _, code := range r.codes {
				if !seen[code] {
					seen[code] = true
					out = append(out, code)
				}
			}
		}
	}
	if len(out) == 0 {
		out = []string{"VL2"}
	}
	return out
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
