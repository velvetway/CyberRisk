package domain

import "time"

// =====================
// ENUM-like types
// =====================

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleAuditor UserRole = "auditor"
	UserRoleViewer  UserRole = "viewer"
)

type AssetEnvironment string

const (
	AssetEnvProd  AssetEnvironment = "prod"
	AssetEnvTest  AssetEnvironment = "test"
	AssetEnvDev   AssetEnvironment = "dev"
	AssetEnvOther AssetEnvironment = "other"
)

type AssetVulnerabilityStatus string

const (
	AssetVulnStatusOpen       AssetVulnerabilityStatus = "open"
	AssetVulnStatusInProgress AssetVulnerabilityStatus = "in_progress"
	AssetVulnStatusMitigated  AssetVulnerabilityStatus = "mitigated"
	AssetVulnStatusAccepted   AssetVulnerabilityStatus = "accepted"
)

type RiskStatus string

const (
	RiskStatusOpen      RiskStatus = "open"
	RiskStatusInReview  RiskStatus = "in_review"
	RiskStatusMitigated RiskStatus = "mitigated"
	RiskStatusAccepted  RiskStatus = "accepted"
)

type RecommendationStatus string

const (
	RecommendationStatusPlanned     RecommendationStatus = "planned"
	RecommendationStatusInProgress  RecommendationStatus = "in_progress"
	RecommendationStatusImplemented RecommendationStatus = "implemented"
	RecommendationStatusRejected    RecommendationStatus = "rejected"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type ThreatSourceType string

const (
	ThreatSourceExternal   ThreatSourceType = "external"
	ThreatSourceInternal   ThreatSourceType = "internal"
	ThreatSourceThirdParty ThreatSourceType = "third_party"
)

// =====================
// Регуляторные категории (152-ФЗ, 187-ФЗ)
// =====================

// DataCategory — категория обрабатываемых данных
type DataCategory string

const (
	DataCategoryPublic             DataCategory = "public"              // Общедоступные
	DataCategoryInternal           DataCategory = "internal"            // Внутренние
	DataCategoryConfidential       DataCategory = "confidential"        // Конфиденциальные
	DataCategoryPersonalData       DataCategory = "personal_data"       // ПДн (152-ФЗ)
	DataCategoryPersonalDataSpec   DataCategory = "personal_data_special" // Специальные ПДн
	DataCategoryPersonalDataBio    DataCategory = "personal_data_biometric" // Биометрические ПДн
	DataCategoryKII                DataCategory = "kii"                 // КИИ (187-ФЗ)
	DataCategoryStateSecret        DataCategory = "state_secret"        // Гостайна
	DataCategoryBankingSecret      DataCategory = "banking_secret"      // Банковская тайна
	DataCategoryMedicalSecret      DataCategory = "medical_secret"      // Врачебная тайна
	DataCategoryCommercialSecret   DataCategory = "commercial_secret"   // Коммерческая тайна
)

// ProtectionLevel — уровень защищённости ПДн (Постановление №1119)
type ProtectionLevel string

const (
	ProtectionLevelUZ1 ProtectionLevel = "uz1" // УЗ-1 (максимальный)
	ProtectionLevelUZ2 ProtectionLevel = "uz2" // УЗ-2
	ProtectionLevelUZ3 ProtectionLevel = "uz3" // УЗ-3
	ProtectionLevelUZ4 ProtectionLevel = "uz4" // УЗ-4 (минимальный)
)

// KIICategory — категория значимости объекта КИИ (187-ФЗ)
type KIICategory string

const (
	KIICategoryNone KIICategory = "none" // Не является объектом КИИ
	KIICategoryCat3 KIICategory = "cat3" // 3 категория (минимальная)
	KIICategoryCat2 KIICategory = "cat2" // 2 категория
	KIICategoryCat1 KIICategory = "cat1" // 1 категория (максимальная)
)

// =====================
// Core entities
// =====================

// User соответствует таблице users.
type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         UserRole  `db:"role"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ---- Assets ----

// AssetTypeRef — справочная таблица типов активов (asset_types в БД)
type AssetTypeRef struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// AssetType — тип актива для автоматического расчёта CIA
type AssetType string

const (
	AssetTypeDatabase    AssetType = "database"
	AssetTypeServer      AssetType = "server"
	AssetTypeApplication AssetType = "application"
	AssetTypeNetwork     AssetType = "network"
	AssetTypeWorkstation AssetType = "workstation"
	AssetTypeMobile      AssetType = "mobile"
	AssetTypeIoT         AssetType = "iot"
	AssetTypeCloud       AssetType = "cloud"
)

type Asset struct {
	ID                  int64            `db:"id"`
	Name                string           `db:"name"`
	Type                *string          `db:"type"`          // Тип актива для расчёта CIA (nullable)
	AssetTypeID         *int16           `db:"asset_type_id"`
	Owner               *string          `db:"owner"`
	Description         *string          `db:"description"`
	Location            *string          `db:"location"`      // Расположение актива
	BusinessCriticality int16            `db:"business_criticality"`
	Confidentiality     int16            `db:"confidentiality"`
	Integrity           int16            `db:"integrity"`
	Availability        int16            `db:"availability"`
	Environment         AssetEnvironment `db:"environment"`

	// Регуляторные поля (152-ФЗ, 187-ФЗ)
	DataCategory       *DataCategory    `db:"data_category"`        // Категория данных
	ProtectionLevel    *ProtectionLevel `db:"protection_level"`     // УЗ для ПДн
	KIICategory        *KIICategory     `db:"kii_category"`         // Категория КИИ
	HasPersonalData    bool             `db:"has_personal_data"`    // Обрабатывает ПДн
	PersonalDataVolume *string          `db:"personal_data_volume"` // Объём ПДн
	HasInternetAccess  bool             `db:"has_internet_access"`  // Доступ в интернет
	IsIsolated         bool             `db:"is_isolated"`          // Изолированный сегмент

	// JSONB. На уровне репозитория можно маршалить/размаршалить в map[string]any.
	Tags      []byte    `db:"tags"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ---- Threats ----

type ThreatCategory struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type Threat struct {
	ID               int64            `db:"id"`
	Name             string           `db:"name"`
	ThreatCategoryID *int16           `db:"threat_category_id"`
	SourceType       ThreatSourceType `db:"source_type"`
	Description      *string          `db:"description"`
	BaseLikelihood   int16            `db:"base_likelihood"`

	// PTSZI модель
	QThreat   float64 `db:"q_threat"   json:"q_threat"`
	QSeverity float64 `db:"q_severity" json:"q_severity"`

	// БДУ ФСТЭК
	BDUID                 *string `db:"bdu_id"`                 // УБИ.001, УБИ.002, etc.
	AttackVector          *string `db:"attack_vector"`          // Сетевой, Локальный, Физический
	ImpactConfidentiality bool    `db:"impact_confidentiality"` // Влияет на конфиденциальность
	ImpactIntegrity       bool    `db:"impact_integrity"`       // Влияет на целостность
	ImpactAvailability    bool    `db:"impact_availability"`    // Влияет на доступность

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ---- Vulnerabilities ----

type VulnerabilityCategory struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type Vulnerability struct {
	ID                      int64     `db:"id"`
	Name                    string    `db:"name"`
	VulnerabilityCategoryID *int16    `db:"vulnerability_category_id"`
	Description             *string   `db:"description"`
	Severity                int16     `db:"severity"`
	AffectsAssetTypeID      *int16    `db:"affects_asset_type_id"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
}

type AssetVulnerability struct {
	ID              int64                    `db:"id"`
	AssetID         int64                    `db:"asset_id"`
	VulnerabilityID int64                    `db:"vulnerability_id"`
	Status          AssetVulnerabilityStatus `db:"status"`
	CreatedAt       time.Time                `db:"created_at"`
	UpdatedAt       time.Time                `db:"updated_at"`
}

// ---- Controls ----

type ControlType struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type Control struct {
	ID                  int64     `db:"id"`
	Name                string    `db:"name"`
	ControlTypeID       *int16    `db:"control_type_id"`
	Description         *string   `db:"description"`
	ReducesLikelihoodBy float64   `db:"reduces_likelihood_by"`
	ReducesImpactBy     float64   `db:"reduces_impact_by"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}

type AssetControl struct {
	ID            int64     `db:"id"`
	AssetID       int64     `db:"asset_id"`
	ControlID     int64     `db:"control_id"`
	Effectiveness float64   `db:"effectiveness"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// ---- Risk scenarios ----

type RiskScenario struct {
	ID              int64      `db:"id"`
	AssetID         int64      `db:"asset_id"`
	ThreatID        int64      `db:"threat_id"`
	VulnerabilityID *int64     `db:"vulnerability_id"`
	Title           string     `db:"title"`
	Description     *string    `db:"description"`
	Impact          *float64   `db:"impact"`     // нормированное 1–5
	Likelihood      *float64   `db:"likelihood"` // нормированное 1–5
	RiskScore       *float64   `db:"risk_score"` // ~1–25
	RiskLevel       *RiskLevel `db:"risk_level"`
	Status          RiskStatus `db:"status"`

	CreatedByUserID *int64    `db:"created_by_user_id"`
	UpdatedByUserID *int64    `db:"updated_by_user_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// ---- Recommendations ----

type RecommendationTemplate struct {
	ID                      int64      `db:"id"`
	Code                    string     `db:"code"`
	Title                   string     `db:"title"`
	Description             *string    `db:"description"`
	AssetTypeID             *int16     `db:"asset_type_id"`
	ThreatCategoryID        *int16     `db:"threat_category_id"`
	VulnerabilityCategoryID *int16     `db:"vulnerability_category_id"`
	MinRiskLevel            *RiskLevel `db:"min_risk_level"`
	CreatedAt               time.Time  `db:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at"`
}

type RiskScenarioRecommendation struct {
	ID                       int64                `db:"id"`
	RiskScenarioID           int64                `db:"risk_scenario_id"`
	RecommendationTemplateID int64                `db:"recommendation_template_id"`
	Status                   RecommendationStatus `db:"status"`
	Comment                  *string              `db:"comment"`
	CreatedAt                time.Time            `db:"created_at"`
	UpdatedAt                time.Time            `db:"updated_at"`
}

// ---- Software Catalog (Справочник ПО) ----

type SoftwareCategory struct {
	ID          int16  `db:"id" json:"id"`
	Code        string `db:"code" json:"code"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description,omitempty"`
}

type Software struct {
	ID         int64   `db:"id" json:"id"`
	Name       string  `db:"name" json:"name"`
	Vendor     string  `db:"vendor" json:"vendor"`
	Version    *string `db:"version" json:"version,omitempty"`
	CategoryID *int16  `db:"category_id" json:"category_id,omitempty"`

	// Реестр российского ПО (Минцифры)
	IsRussian      bool    `db:"is_russian" json:"is_russian"`
	RegistryNumber *string `db:"registry_number" json:"registry_number,omitempty"`
	RegistryDate   *string `db:"registry_date" json:"registry_date,omitempty"`
	RegistryURL    *string `db:"registry_url" json:"registry_url,omitempty"`

	// Сертификация ФСТЭК
	FSTECCertified       bool    `db:"fstec_certified" json:"fstec_certified"`
	FSTECCertificateNum  *string `db:"fstec_certificate_num" json:"fstec_certificate_num,omitempty"`
	FSTECCertificateDate *string `db:"fstec_certificate_date" json:"fstec_certificate_date,omitempty"`
	FSTECProtectionClass *string `db:"fstec_protection_class" json:"fstec_protection_class,omitempty"`
	FSTECValidUntil      *string `db:"fstec_valid_until" json:"fstec_valid_until,omitempty"`

	// Сертификация ФСБ (СКЗИ)
	FSBCertified       bool    `db:"fsb_certified" json:"fsb_certified"`
	FSBCertificateNum  *string `db:"fsb_certificate_num" json:"fsb_certificate_num,omitempty"`
	FSBProtectionClass *string `db:"fsb_protection_class" json:"fsb_protection_class,omitempty"` // КС1, КС2, КС3, КВ, КА

	Description *string   `db:"description" json:"description,omitempty"`
	Website     *string   `db:"website" json:"website,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type AssetSoftware struct {
	ID             int64     `db:"id"`
	AssetID        int64     `db:"asset_id"`
	SoftwareID     int64     `db:"software_id"`
	Version        *string   `db:"version"`
	InstallDate    *string   `db:"install_date"`
	LicenseType    *string   `db:"license_type"`
	LicenseExpires *string   `db:"license_expires"`
	Notes          *string   `db:"notes"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// SoftwareWithCategory — ПО с категорией для отображения
type SoftwareWithCategory struct {
	Software
	CategoryName string `db:"category_name" json:"category_name,omitempty"`
}

// AssetSoftwareWithSoftware — запись о ПО на активе вместе с карточкой из справочника
type AssetSoftwareWithSoftware struct {
	Link     AssetSoftware `json:"link"`
	Software Software      `json:"software"`
}

// =====================
// Вспомогательные структуры
// =====================

// AssetControlWithControl — удобный агрегат для калькуляции риска:
// одна запись = связка "мера защиты на активе" + описание самой меры.
type AssetControlWithControl struct {
	AssetControl AssetControl
	Control      Control
}

// RiskCalculationContext — все данные, которые нужны для расчёта одного сценария риска.
type RiskCalculationContext struct {
	Asset    Asset
	Threat   Threat
	Vuln     *Vulnerability
	Controls []AssetControlWithControl
	Scenario *RiskScenario
}

// RiskCalculator — интерфейс для сервиса расчёта риска.
// Реализация будет использовать нашу формулу:
// Impact, Likelihood, RiskScore, RiskLevel.
type RiskCalculator interface {
	Calculate(ctx RiskCalculationContext) error
}
