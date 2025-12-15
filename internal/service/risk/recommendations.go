// internal/service/risk/recommendations.go
package risk

import (
	"strings"

	"Diplom/internal/domain"
)

// RiskRecommendation описывает одно текстовое предложение по снижению риска.
type RiskRecommendation struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Priority: "low" | "medium" | "high"
	Priority string `json:"priority"`
	// Category: high-level группа
	Category string `json:"category"`
}

// RussianTool описывает рекомендуемое российское СЗИ
type RussianTool struct {
	Name            string `json:"name"`
	Vendor          string `json:"vendor"`
	Description     string `json:"description"`
	FSTECCertified  bool   `json:"fstec_certified"`
	FSBCertified    bool   `json:"fsb_certified"`
	RegistryNumber  string `json:"registry_number"` // Номер в реестре Минцифры
	ProtectionClass string `json:"protection_class"`
	UseCase         string `json:"use_case"`
}

// RegulatoryRecommendation — рекомендация по соответствию регуляторным требованиям
type RegulatoryRecommendation struct {
	RiskRecommendation
	Regulation    string        `json:"regulation"`     // 152-ФЗ, 187-ФЗ и т.д.
	Requirement   string        `json:"requirement"`    // Конкретное требование
	RussianTools  []RussianTool `json:"russian_tools"`  // Рекомендуемые российские СЗИ
	Deadline      string        `json:"deadline"`       // Срок выполнения (рекомендуемый)
	Penalty       string        `json:"penalty"`        // Санкции за невыполнение
}

// Каталог российских СЗИ
var russianSecurityTools = map[string][]RussianTool{
	"antivirus": {
		{
			Name:            "Kaspersky Endpoint Security",
			Vendor:          "Лаборатория Касперского",
			Description:     "Комплексная защита конечных точек от вредоносного ПО",
			FSTECCertified:  true,
			RegistryNumber:  "55",
			ProtectionClass: "4",
			UseCase:         "Защита от вредоносного ПО, контроль устройств, веб-контроль",
		},
		{
			Name:            "Dr.Web Enterprise Security Suite",
			Vendor:          "Доктор Веб",
			Description:     "Антивирусная защита для корпоративных сетей",
			FSTECCertified:  true,
			RegistryNumber:  "66",
			ProtectionClass: "4",
			UseCase:         "Антивирусная защита серверов и рабочих станций",
		},
	},
	"access_control": {
		{
			Name:            "Secret Net Studio",
			Vendor:          "Код Безопасности",
			Description:     "Сертифицированное СЗИ от НСД для защиты рабочих станций и серверов",
			FSTECCertified:  true,
			RegistryNumber:  "83",
			ProtectionClass: "4",
			UseCase:         "Контроль доступа, разграничение полномочий, контроль целостности",
		},
		{
			Name:            "Dallas Lock",
			Vendor:          "Конфидент",
			Description:     "СЗИ от НСД с поддержкой электронных замков",
			FSTECCertified:  true,
			RegistryNumber:  "104",
			ProtectionClass: "4",
			UseCase:         "Контроль доступа, двухфакторная аутентификация",
		},
		{
			Name:            "Аккорд-АМДЗ",
			Vendor:          "ОКБ САПР",
			Description:     "Аппаратно-программный модуль доверенной загрузки",
			FSTECCertified:  true,
			RegistryNumber:  "17",
			ProtectionClass: "3",
			UseCase:         "Доверенная загрузка, контроль целостности на этапе загрузки",
		},
	},
	"cryptography": {
		{
			Name:            "КриптоПро CSP",
			Vendor:          "КриптоПро",
			Description:     "Сертифицированное СКЗИ для криптографической защиты информации",
			FSTECCertified:  true,
			FSBCertified:    true,
			RegistryNumber:  "124",
			ProtectionClass: "КС3",
			UseCase:         "Шифрование данных, ЭЦП, защита каналов связи",
		},
		{
			Name:            "ViPNet CSP",
			Vendor:          "ИнфоТеКС",
			Description:     "Криптопровайдер для защиты информации",
			FSTECCertified:  true,
			FSBCertified:    true,
			RegistryNumber:  "76",
			ProtectionClass: "КС2",
			UseCase:         "Шифрование, ЭЦП, защита персональных данных",
		},
	},
	"network_security": {
		{
			Name:            "ViPNet Coordinator",
			Vendor:          "ИнфоТеКС",
			Description:     "VPN-шлюз и межсетевой экран",
			FSTECCertified:  true,
			FSBCertified:    true,
			RegistryNumber:  "76",
			ProtectionClass: "3",
			UseCase:         "Защита периметра, VPN, межсетевое экранирование",
		},
		{
			Name:            "Континент",
			Vendor:          "Код Безопасности",
			Description:     "Аппаратно-программный комплекс защиты сетей",
			FSTECCertified:  true,
			FSBCertified:    true,
			RegistryNumber:  "83",
			ProtectionClass: "3",
			UseCase:         "NGFW, VPN, обнаружение вторжений",
		},
		{
			Name:            "UserGate",
			Vendor:          "UserGate",
			Description:     "Межсетевой экран нового поколения",
			FSTECCertified:  true,
			RegistryNumber:  "3905",
			ProtectionClass: "4",
			UseCase:         "NGFW, контентная фильтрация, IDS/IPS",
		},
	},
	"siem": {
		{
			Name:            "MaxPatrol SIEM",
			Vendor:          "Positive Technologies",
			Description:     "Система управления событиями информационной безопасности",
			FSTECCertified:  true,
			RegistryNumber:  "1143",
			ProtectionClass: "4",
			UseCase:         "Корреляция событий ИБ, выявление инцидентов, compliance",
		},
		{
			Name:            "ViPNet TIAS",
			Vendor:          "ИнфоТеКС",
			Description:     "Система анализа событий информационной безопасности",
			FSTECCertified:  true,
			RegistryNumber:  "76",
			ProtectionClass: "4",
			UseCase:         "Сбор и анализ событий ИБ",
		},
		{
			Name:            "RuSIEM",
			Vendor:          "RuSIEM",
			Description:     "Российская SIEM-система",
			FSTECCertified:  true,
			RegistryNumber:  "6604",
			ProtectionClass: "4",
			UseCase:         "Мониторинг и анализ событий ИБ",
		},
	},
	"waf": {
		{
			Name:            "PT Application Firewall",
			Vendor:          "Positive Technologies",
			Description:     "Межсетевой экран уровня веб-приложений",
			FSTECCertified:  true,
			RegistryNumber:  "1143",
			ProtectionClass: "4",
			UseCase:         "Защита веб-приложений от атак OWASP Top 10",
		},
		{
			Name:            "Wallarm",
			Vendor:          "Wallarm",
			Description:     "Платформа защиты веб-приложений и API",
			FSTECCertified:  false,
			RegistryNumber:  "6580",
			UseCase:         "WAF, защита API, обнаружение уязвимостей",
		},
	},
	"backup": {
		{
			Name:            "Кибер Бэкап",
			Vendor:          "Киберпротект",
			Description:     "Система резервного копирования и восстановления",
			FSTECCertified:  true,
			RegistryNumber:  "489",
			ProtectionClass: "4",
			UseCase:         "Резервное копирование, защита от ransomware",
		},
		{
			Name:            "RuBackup",
			Vendor:          "Рубэкап",
			Description:     "Российская система резервного копирования",
			FSTECCertified:  false,
			RegistryNumber:  "6078",
			UseCase:         "Резервное копирование корпоративных данных",
		},
	},
	"dlp": {
		{
			Name:            "InfoWatch Traffic Monitor",
			Vendor:          "InfoWatch",
			Description:     "DLP-система для защиты от утечек данных",
			FSTECCertified:  true,
			RegistryNumber:  "205",
			ProtectionClass: "4",
			UseCase:         "Предотвращение утечек данных, контроль каналов передачи",
		},
		{
			Name:            "СёрчИнформ КИБ",
			Vendor:          "СёрчИнформ",
			Description:     "Комплексная система защиты от внутренних угроз",
			FSTECCertified:  true,
			RegistryNumber:  "3177",
			ProtectionClass: "4",
			UseCase:         "DLP, контроль рабочего времени, расследование инцидентов",
		},
	},
	"vulnerability_scanner": {
		{
			Name:            "MaxPatrol 8",
			Vendor:          "Positive Technologies",
			Description:     "Система контроля защищённости и соответствия стандартам",
			FSTECCertified:  true,
			RegistryNumber:  "1143",
			ProtectionClass: "4",
			UseCase:         "Сканирование уязвимостей, compliance, инвентаризация",
		},
		{
			Name:            "RedCheck",
			Vendor:          "АЛТЭКС-СОФТ",
			Description:     "Сканер уязвимостей с поддержкой БДУ ФСТЭК",
			FSTECCertified:  true,
			RegistryNumber:  "1164",
			ProtectionClass: "4",
			UseCase:         "Сканирование уязвимостей, контроль конфигураций",
		},
		{
			Name:            "XSpider",
			Vendor:          "Positive Technologies",
			Description:     "Сканер безопасности сетей",
			FSTECCertified:  true,
			RegistryNumber:  "1143",
			ProtectionClass: "4",
			UseCase:         "Сетевой аудит, поиск уязвимостей",
		},
	},
}

// GenerateRecommendations — rule-based движок рекомендаций с учётом
// российских регуляторных требований и сертифицированных СЗИ.
func GenerateRecommendations(
	asset *domain.Asset,
	threat *domain.Threat,
	vulns []domain.Vulnerability,
	res RiskResult,
) []RiskRecommendation {
	var out []RiskRecommendation

	add := func(code, title, desc, priority, category string) {
		out = append(out, RiskRecommendation{
			Code:        code,
			Title:       title,
			Description: desc,
			Priority:    priority,
			Category:    category,
		})
	}

	level := strings.ToLower(res.Level)

	// 1) Общие рекомендации по уровню риска
	switch level {
	case "critical":
		add(
			"RISK_GENERAL_CRITICAL",
			"Снизить критический риск до приемлемого уровня",
			"Необходимо в кратчайшие сроки разработать и внедрить план обработки риска: технические меры, организационные ограничения, возможная приостановка сервисов до устранения критичных уязвимостей.",
			"high",
			"general",
		)
	case "high":
		add(
			"RISK_GENERAL_HIGH",
			"Приоритизировать данный риск в плане мероприятий",
			"Риск высокого уровня должен войти в приоритетный план мероприятий по ИБ с указанием сроков, ответственных и бюджета.",
			"high",
			"general",
		)
	case "medium":
		add(
			"RISK_GENERAL_MEDIUM",
			"Запланировать снижение риска до низкого уровня",
			"Рекомендуется включить данный риск в план работ по ИБ на очередной период и предусмотреть меры для снижения вероятности или воздействия.",
			"medium",
			"general",
		)
	default:
		add(
			"RISK_GENERAL_LOW",
			"Поддерживать текущий уровень контроля",
			"Риск находится на низком уровне. Важно поддерживать действующие меры защиты и периодически пересматривать актуальность угроз и уязвимостей.",
			"low",
			"general",
		)
	}

	// 2) Рекомендации по регуляторным требованиям
	generateRegulatoryRecommendations(asset, &out, level)

	// 3) Особенности среды (production)
	if strings.EqualFold(string(asset.Environment), "prod") {
		add(
			"ASSET_ENV_PROD_HARDENING",
			"Усиление защиты в продуктивной среде",
			"Актив находится в продуктивной среде. Рекомендуется минимизировать доступы, применять принцип наименьших привилегий, ограничивать администраторский доступ и сегментировать сеть.",
			"high",
			"hardening",
		)
	}

	// 4) Рекомендации по типу угрозы
	generateThreatRecommendations(threat, &out)

	// 5) Рекомендации по БДУ ФСТЭК
	generateBDURecommendations(threat, asset, &out)

	// 6) Уязвимости: наличие критичных / высоких по Severity
	generateVulnerabilityRecommendations(vulns, &out)

	// 7) Защита конфиденциальных данных
	if asset.Confidentiality >= 4 && (level == "high" || level == "critical") {
		add(
			"ASSET_CONF_ENCRYPTION",
			"Защита конфиденциальных данных",
			"Рекомендуется применить шифрование данных «на диске» и «в канале» с использованием сертифицированных СКЗИ (КриптоПро CSP, ViPNet CSP), а также усилить контроль доступа к данным.",
			"high",
			"data_protection",
		)
	}

	// 8) Рекомендации по изоляции сети
	if asset.HasInternetAccess && asset.Confidentiality >= 4 {
		add(
			"ASSET_INTERNET_ISOLATION",
			"Рассмотреть изоляцию критичного актива от интернета",
			"Актив с высокой конфиденциальностью имеет доступ в интернет. Рекомендуется рассмотреть сетевую изоляцию или применение DMZ с сертифицированными межсетевыми экранами (ViPNet Coordinator, Континент, UserGate).",
			"high",
			"network_security",
		)
	}

	return out
}

// generateRegulatoryRecommendations добавляет рекомендации по соответствию регуляторным требованиям
func generateRegulatoryRecommendations(asset *domain.Asset, out *[]RiskRecommendation, level string) {
	add := func(code, title, desc, priority, category string) {
		*out = append(*out, RiskRecommendation{
			Code:        code,
			Title:       title,
			Description: desc,
			Priority:    priority,
			Category:    category,
		})
	}

	// КИИ (187-ФЗ)
	if asset.KIICategory != nil && *asset.KIICategory != domain.KIICategoryNone {
		category := *asset.KIICategory
		switch category {
		case domain.KIICategoryCat1:
			add(
				"KII_CAT1_COMPLIANCE",
				"Обеспечить соответствие требованиям КИИ 1 категории (187-ФЗ)",
				"Объект КИИ 1 категории значимости требует максимального уровня защиты. Необходимо: "+
					"1) Внедрить сертифицированные СЗИ класса не ниже 3 (Secret Net Studio, Dallas Lock); "+
					"2) Организовать круглосуточный мониторинг с использованием SIEM (MaxPatrol SIEM); "+
					"3) Обеспечить подключение к ГосСОПКА; "+
					"4) Провести аттестацию объекта информатизации.",
				"high",
				"regulatory_compliance",
			)
		case domain.KIICategoryCat2:
			add(
				"KII_CAT2_COMPLIANCE",
				"Обеспечить соответствие требованиям КИИ 2 категории (187-ФЗ)",
				"Объект КИИ 2 категории требует усиленной защиты. Необходимо: "+
					"1) Внедрить сертифицированные СЗИ от НСД; "+
					"2) Организовать мониторинг событий ИБ; "+
					"3) Обеспечить взаимодействие с ГосСОПКА; "+
					"4) Разработать и утвердить модель угроз.",
				"high",
				"regulatory_compliance",
			)
		case domain.KIICategoryCat3:
			add(
				"KII_CAT3_COMPLIANCE",
				"Обеспечить соответствие требованиям КИИ 3 категории (187-ФЗ)",
				"Объект КИИ 3 категории требует базовой защиты. Необходимо: "+
					"1) Внедрить СЗИ в соответствии с приказом ФСТЭК №239; "+
					"2) Организовать регистрацию событий безопасности; "+
					"3) Обеспечить резервное копирование критичных данных.",
				"medium",
				"regulatory_compliance",
			)
		}

		// Общие рекомендации для КИИ
		add(
			"KII_GOSSOPKA",
			"Обеспечить подключение к ГосСОПКА",
			"В соответствии с 187-ФЗ субъекты КИИ обязаны взаимодействовать с ГосСОПКА для обмена информацией о компьютерных инцидентах. "+
				"Рекомендуется использовать сертифицированные средства защиты с поддержкой интеграции: ViPNet TIAS, MaxPatrol SIEM.",
			"high",
			"regulatory_compliance",
		)

		add(
			"KII_IMPORT_SUBSTITUTION",
			"Провести импортозамещение в соответствии с требованиями для КИИ",
			"Указ Президента №166 от 30.03.2022 запрещает использование иностранного ПО на значимых объектах КИИ с 01.01.2025. "+
				"Рекомендуется переход на российские ОС (Astra Linux, Альт), СУБД (Postgres Pro), офисное ПО (МойОфис).",
			"high",
			"regulatory_compliance",
		)
	}

	// ПДн (152-ФЗ)
	if asset.HasPersonalData {
		priority := "medium"
		if asset.ProtectionLevel != nil {
			switch *asset.ProtectionLevel {
			case domain.ProtectionLevelUZ1, domain.ProtectionLevelUZ2:
				priority = "high"
			}
		}

		add(
			"PDN_152FZ_COMPLIANCE",
			"Обеспечить соответствие требованиям 152-ФЗ",
			"Актив обрабатывает персональные данные. Необходимо: "+
				"1) Определить уровень защищённости ПДн в соответствии с ПП-1119; "+
				"2) Внедрить СЗИ согласно приказу ФСТЭК №21; "+
				"3) Разработать модель угроз; "+
				"4) Уведомить Роскомнадзор об обработке ПДн.",
			priority,
			"regulatory_compliance",
		)

		if asset.ProtectionLevel != nil {
			switch *asset.ProtectionLevel {
			case domain.ProtectionLevelUZ1:
				add(
					"PDN_UZ1_REQUIREMENTS",
					"Выполнить требования для УЗ-1 (максимальный уровень защищённости ПДн)",
					"УЗ-1 требует: "+
						"1) СЗИ от НСД класса не ниже 4 (Secret Net Studio, Dallas Lock); "+
						"2) СКЗИ класса не ниже КС3 (КриптоПро CSP); "+
						"3) Средства обнаружения вторжений класса 4; "+
						"4) Антивирусные средства класса 4 (Kaspersky, Dr.Web); "+
						"5) Межсетевое экранирование класса 4 (Континент, UserGate).",
					"high",
					"regulatory_compliance",
				)
			case domain.ProtectionLevelUZ2:
				add(
					"PDN_UZ2_REQUIREMENTS",
					"Выполнить требования для УЗ-2",
					"УЗ-2 требует: "+
						"1) СЗИ от НСД класса не ниже 5; "+
						"2) СКЗИ класса не ниже КС2; "+
						"3) Средства обнаружения вторжений класса 5; "+
						"4) Антивирусные средства класса 5.",
					"high",
					"regulatory_compliance",
				)
			}
		}

		// Объём ПДн
		if asset.PersonalDataVolume != nil && *asset.PersonalDataVolume == ">100000" {
			add(
				"PDN_VOLUME_HIGH",
				"Усиленные меры защиты для большого объёма ПДн",
				"Обработка ПДн более 100 000 субъектов требует дополнительных мер: "+
					"1) Назначение ответственного за обработку ПДн; "+
					"2) Регулярный аудит процессов обработки; "+
					"3) DLP-система для контроля утечек (InfoWatch, СёрчИнформ КИБ).",
				"high",
				"regulatory_compliance",
			)
		}
	}

	// Категории данных
	if asset.DataCategory != nil {
		switch *asset.DataCategory {
		case domain.DataCategoryStateSecret:
			add(
				"DATA_STATE_SECRET",
				"Обеспечить защиту государственной тайны",
				"Актив содержит сведения, составляющие государственную тайну. Требуется: "+
					"1) Аттестация объекта информатизации; "+
					"2) Использование СКЗИ класса КВ/КА; "+
					"3) Специальные проверки помещений; "+
					"4) Режим секретности в соответствии с Законом о гостайне.",
				"high",
				"regulatory_compliance",
			)
		case domain.DataCategoryBankingSecret:
			add(
				"DATA_BANKING_SECRET",
				"Обеспечить защиту банковской тайны",
				"Актив содержит сведения, составляющие банковскую тайну (395-1-ФЗ). Требуется: "+
					"1) Соответствие требованиям ЦБ РФ (ГОСТ 57580.1); "+
					"2) Шифрование данных в соответствии с СТО БР ИББС; "+
					"3) Регулярное тестирование на проникновение.",
				"high",
				"regulatory_compliance",
			)
		case domain.DataCategoryMedicalSecret:
			add(
				"DATA_MEDICAL_SECRET",
				"Обеспечить защиту врачебной тайны",
				"Актив содержит сведения, составляющие врачебную тайну (323-ФЗ). Требуется: "+
					"1) Разграничение доступа по ролям медперсонала; "+
					"2) Журналирование всех обращений к медицинским данным; "+
					"3) Шифрование при передаче между учреждениями.",
				"high",
				"regulatory_compliance",
			)
		case domain.DataCategoryPersonalDataSpec, domain.DataCategoryPersonalDataBio:
			add(
				"DATA_PDN_SPECIAL",
				"Усиленная защита специальных/биометрических ПДн",
				"Обработка специальных категорий ПДн или биометрических данных требует: "+
					"1) Письменное согласие субъекта на обработку; "+
					"2) Повышенный уровень защищённости (УЗ-1 или УЗ-2); "+
					"3) Шифрование при хранении и передаче.",
				"high",
				"regulatory_compliance",
			)
		}
	}
}

// generateThreatRecommendations добавляет рекомендации по типу угрозы
func generateThreatRecommendations(threat *domain.Threat, out *[]RiskRecommendation) {
	add := func(code, title, desc, priority, category string) {
		*out = append(*out, RiskRecommendation{
			Code:        code,
			Title:       title,
			Description: desc,
			Priority:    priority,
			Category:    category,
		})
	}

	threatName := strings.ToLower(threat.Name)

	if strings.Contains(threatName, "brute") || strings.Contains(threatName, "password") ||
		strings.Contains(threatName, "подбор") || strings.Contains(threatName, "пароль") {
		add(
			"THREAT_BRUTE_FORCE",
			"Защита от подбора учётных данных",
			"Рекомендуется: "+
				"1) Включить политику блокировки учётных записей; "+
				"2) Внедрить многофакторную аутентификацию (JaCarta, Рутокен); "+
				"3) Использовать СЗИ от НСД с контролем паролей (Secret Net Studio).",
			"high",
			"access_control",
		)
	}

	if strings.Contains(threatName, "sql injection") || strings.Contains(threatName, "injection") ||
		strings.Contains(threatName, "инъекция") || strings.Contains(threatName, "внедрение") {
		add(
			"THREAT_SQL_INJECTION",
			"Защита от SQL-инъекций и внедрения данных",
			"Рекомендуется: "+
				"1) Использовать параметризованные запросы; "+
				"2) Внедрить WAF (PT Application Firewall); "+
				"3) Проводить регулярный анализ кода (PT Application Inspector).",
			"high",
			"application_security",
		)
	}

	if strings.Contains(threatName, "ransom") || strings.Contains(threatName, "ransomware") ||
		strings.Contains(threatName, "шифровальщик") || strings.Contains(threatName, "вымогатель") {
		add(
			"THREAT_RANSOMWARE",
			"Защита от программ-вымогателей",
			"Рекомендуется: "+
				"1) Регулярное резервное копирование с изоляцией копий (Кибер Бэкап); "+
				"2) Антивирусная защита с поведенческим анализом (Kaspersky EDR); "+
				"3) Ограничение прав пользователей; "+
				"4) Контроль запуска приложений (Secret Net Studio).",
			"high",
			"backup",
		)
	}

	if strings.Contains(threatName, "ddos") || strings.Contains(threatName, "отказ в обслуживании") {
		add(
			"THREAT_DDOS",
			"Защита от DDoS-атак",
			"Рекомендуется: "+
				"1) Использовать сервисы защиты от DDoS (Qrator, DDoS-Guard); "+
				"2) Настроить rate limiting на WAF; "+
				"3) Обеспечить резервирование каналов связи.",
			"high",
			"network_security",
		)
	}

	if strings.Contains(threatName, "утечка") || strings.Contains(threatName, "leakage") ||
		strings.Contains(threatName, "инсайдер") || strings.Contains(threatName, "insider") {
		add(
			"THREAT_DATA_LEAKAGE",
			"Защита от утечки данных",
			"Рекомендуется: "+
				"1) Внедрить DLP-систему (InfoWatch Traffic Monitor, СёрчИнформ КИБ); "+
				"2) Классифицировать данные и установить метки; "+
				"3) Контролировать съёмные носители; "+
				"4) Организовать мониторинг действий привилегированных пользователей.",
			"high",
			"data_protection",
		)
	}
}

// generateBDURecommendations добавляет рекомендации на основе БДУ ФСТЭК
func generateBDURecommendations(threat *domain.Threat, asset *domain.Asset, out *[]RiskRecommendation) {
	if threat.BDUID == nil || *threat.BDUID == "" {
		return
	}

	add := func(code, title, desc, priority, category string) {
		*out = append(*out, RiskRecommendation{
			Code:        code,
			Title:       title,
			Description: desc,
			Priority:    priority,
			Category:    category,
		})
	}

	// Рекомендации по вектору атаки
	if threat.AttackVector != nil {
		vector := strings.ToLower(*threat.AttackVector)
		if strings.Contains(vector, "сетевой") || strings.Contains(vector, "network") {
			add(
				"BDU_NETWORK_VECTOR",
				"Защита от сетевых атак (БДУ ФСТЭК)",
				"Угроза реализуется через сетевой вектор. Рекомендуется: "+
					"1) Сегментация сети с использованием VLAN; "+
					"2) Межсетевое экранирование (Континент, UserGate, ViPNet Coordinator); "+
					"3) Система обнаружения вторжений (ViPNet IDS, Континент IDS).",
				"high",
				"network_security",
			)
		}
		if strings.Contains(vector, "локальный") || strings.Contains(vector, "local") {
			add(
				"BDU_LOCAL_VECTOR",
				"Защита от локальных атак (БДУ ФСТЭК)",
				"Угроза реализуется локально. Рекомендуется: "+
					"1) СЗИ от НСД (Secret Net Studio, Dallas Lock); "+
					"2) Контроль целостности файлов и конфигураций; "+
					"3) Ограничение административных привилегий.",
				"high",
				"access_control",
			)
		}
	}

	// Рекомендации по влиянию на CIA
	if threat.ImpactConfidentiality && asset.Confidentiality >= 4 {
		add(
			"BDU_IMPACT_CONF",
			"Защита конфиденциальности от угрозы БДУ",
			"Угроза влияет на конфиденциальность критичного актива. Рекомендуется: "+
				"1) Шифрование данных (КриптоПро CSP, ViPNet CSP); "+
				"2) DLP-система (InfoWatch, СёрчИнформ); "+
				"3) Контроль доступа и журналирование.",
			"high",
			"data_protection",
		)
	}

	if threat.ImpactIntegrity && asset.Integrity >= 4 {
		add(
			"BDU_IMPACT_INT",
			"Защита целостности от угрозы БДУ",
			"Угроза влияет на целостность критичного актива. Рекомендуется: "+
				"1) Контроль целостности файлов (Secret Net Studio); "+
				"2) Резервное копирование с проверкой (Кибер Бэкап); "+
				"3) Цифровая подпись критичных данных.",
			"high",
			"data_protection",
		)
	}

	if threat.ImpactAvailability && asset.Availability >= 4 {
		add(
			"BDU_IMPACT_AVAIL",
			"Защита доступности от угрозы БДУ",
			"Угроза влияет на доступность критичного актива. Рекомендуется: "+
				"1) Резервирование критичных компонентов; "+
				"2) План обеспечения непрерывности (BCP/DRP); "+
				"3) Регулярное тестирование восстановления.",
			"high",
			"availability",
		)
	}
}

// generateVulnerabilityRecommendations добавляет рекомендации по уязвимостям
func generateVulnerabilityRecommendations(vulns []domain.Vulnerability, out *[]RiskRecommendation) {
	var maxSeverity int16 = 0
	for i := range vulns {
		if vulns[i].Severity > maxSeverity {
			maxSeverity = vulns[i].Severity
		}
	}

	add := func(code, title, desc, priority, category string) {
		*out = append(*out, RiskRecommendation{
			Code:        code,
			Title:       title,
			Description: desc,
			Priority:    priority,
			Category:    category,
		})
	}

	if maxSeverity >= 8 {
		add(
			"VULN_CRITICAL_PATCHING",
			"Оперативное устранение критичных уязвимостей",
			"Обнаружены уязвимости высокого/критического уровня. Рекомендуется: "+
				"1) Оперативно установить исправляющие обновления; "+
				"2) При невозможности патчинга — применить компенсирующие меры; "+
				"3) Использовать сканер уязвимостей для контроля (MaxPatrol 8, RedCheck).",
			"high",
			"vulnerability_management",
		)
	} else if maxSeverity >= 5 {
		add(
			"VULN_MEDIUM_PATCHING",
			"Плановое устранение уязвимостей среднего уровня",
			"Рекомендуется включить устранение выявленных уязвимостей в плановое окно обслуживания. "+
				"Для мониторинга состояния рекомендуется использовать MaxPatrol 8 или RedCheck.",
			"medium",
			"vulnerability_management",
		)
	}

	if len(vulns) > 5 {
		add(
			"VULN_PROCESS_IMPROVEMENT",
			"Улучшить процесс управления уязвимостями",
			"Обнаружено значительное количество уязвимостей. Рекомендуется: "+
				"1) Внедрить автоматизированный процесс сканирования (MaxPatrol 8); "+
				"2) Установить SLA на устранение уязвимостей по критичности; "+
				"3) Интегрировать данные об уязвимостях в SIEM.",
			"medium",
			"vulnerability_management",
		)
	}
}

// GenerateRegulatoryRecommendations генерирует расширенные рекомендации
// с информацией о российских СЗИ для конкретного актива
func GenerateRegulatoryRecommendations(asset *domain.Asset) []RegulatoryRecommendation {
	var recs []RegulatoryRecommendation

	// КИИ (187-ФЗ)
	if asset.KIICategory != nil && *asset.KIICategory != domain.KIICategoryNone {
		rec := RegulatoryRecommendation{
			RiskRecommendation: RiskRecommendation{
				Code:        "KII_187FZ",
				Title:       "Соответствие требованиям 187-ФЗ (КИИ)",
				Description: "Необходимо обеспечить защиту значимого объекта КИИ в соответствии с приказами ФСТЭК №235 и №239.",
				Priority:    "high",
				Category:    "regulatory_compliance",
			},
			Regulation:  "187-ФЗ",
			Requirement: "Приказ ФСТЭК России №239",
			Deadline:    "В соответствии с категорией значимости",
			Penalty:     "УК РФ ст. 274.1 — до 10 лет лишения свободы",
		}

		// Добавляем рекомендуемые СЗИ в зависимости от категории
		rec.RussianTools = append(rec.RussianTools, russianSecurityTools["access_control"]...)
		rec.RussianTools = append(rec.RussianTools, russianSecurityTools["siem"]...)
		rec.RussianTools = append(rec.RussianTools, russianSecurityTools["network_security"]...)

		recs = append(recs, rec)
	}

	// ПДн (152-ФЗ)
	if asset.HasPersonalData {
		rec := RegulatoryRecommendation{
			RiskRecommendation: RiskRecommendation{
				Code:        "PDN_152FZ",
				Title:       "Соответствие требованиям 152-ФЗ (ПДн)",
				Description: "Необходимо обеспечить защиту персональных данных в соответствии с требованиями ПП-1119 и приказа ФСТЭК №21.",
				Priority:    "high",
				Category:    "regulatory_compliance",
			},
			Regulation:  "152-ФЗ",
			Requirement: "ПП-1119, Приказ ФСТЭК России №21",
			Deadline:    "До начала обработки ПДн",
			Penalty:     "КоАП РФ ст. 13.11 — штраф до 18 млн руб.",
		}

		rec.RussianTools = append(rec.RussianTools, russianSecurityTools["access_control"]...)
		rec.RussianTools = append(rec.RussianTools, russianSecurityTools["antivirus"]...)
		if asset.ProtectionLevel != nil && (*asset.ProtectionLevel == domain.ProtectionLevelUZ1 || *asset.ProtectionLevel == domain.ProtectionLevelUZ2) {
			rec.RussianTools = append(rec.RussianTools, russianSecurityTools["cryptography"]...)
		}

		recs = append(recs, rec)
	}

	// Защита от утечек
	if asset.Confidentiality >= 4 {
		rec := RegulatoryRecommendation{
			RiskRecommendation: RiskRecommendation{
				Code:        "DLP_CONF",
				Title:       "Внедрение DLP-системы для защиты конфиденциальных данных",
				Description: "Для актива с высоким уровнем конфиденциальности рекомендуется внедрить систему предотвращения утечек данных.",
				Priority:    "high",
				Category:    "data_protection",
			},
			Regulation:  "Внутренние требования ИБ",
			Requirement: "Контроль каналов передачи данных",
		}
		rec.RussianTools = russianSecurityTools["dlp"]

		recs = append(recs, rec)
	}

	return recs
}

// GetRussianToolsByCategory возвращает список российских СЗИ по категории
func GetRussianToolsByCategory(category string) []RussianTool {
	if tools, ok := russianSecurityTools[category]; ok {
		return tools
	}
	return nil
}

// GetAllRussianTools возвращает все российские СЗИ из каталога
func GetAllRussianTools() map[string][]RussianTool {
	return russianSecurityTools
}
