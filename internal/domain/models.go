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

type ThreatSourceType string

const (
	ThreatSourceExternal   ThreatSourceType = "external"
	ThreatSourceInternal   ThreatSourceType = "internal"
	ThreatSourceThirdParty ThreatSourceType = "third_party"
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

// AssetTypeRef — справочная таблица типов активов (asset_types в БД).
type AssetTypeRef struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// Asset — актив под PTSZI W-моделью. Только поля, которые либо участвуют в W
// (`is_isolated`), либо нужны UI/отчётам для маркировки. Legacy-поля
// (C/I/A, business_criticality, kii_category, data_category, protection_level,
// has_personal_data, personal_data_volume, has_internet_access, type, location)
// удалены миграцией 020 — см. docs/risk-model.md.
type Asset struct {
	ID          int64            `db:"id"`
	Name        string           `db:"name"`
	AssetTypeID *int16           `db:"asset_type_id"`
	Owner       *string          `db:"owner"`
	Description *string          `db:"description"`
	Environment AssetEnvironment `db:"environment"`
	IsIsolated  bool             `db:"is_isolated"`

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

// Threat — угроза в PTSZI-модели. q_threat и q_severity напрямую кормят формулу W.
//
// AppliesToAssetTypes — список id типов активов, на которые угроза реально
// нацелена (выводится импортёром из «Объект воздействия» ФСТЭК). Пустой
// массив = «применима ко всем типам». Используется applicability-фильтром
// в risk service: пара (актив, угроза) исключается из overview, если тип
// актива не входит в этот список.
type Threat struct {
	ID               int64            `db:"id"`
	Name             string           `db:"name"`
	ThreatCategoryID *int16           `db:"threat_category_id"`
	SourceType       ThreatSourceType `db:"source_type"`
	Description      *string          `db:"description"`

	QThreat   float64 `db:"q_threat"   json:"q_threat"`
	QSeverity float64 `db:"q_severity" json:"q_severity"`

	BDUID *string `db:"bdu_id"` // УБИ.001, УБИ.002 и т.д.

	AppliesToTargets    *string `db:"applies_to_targets"     json:"applies_to_targets,omitempty"`
	AppliesToAssetTypes []int16 `db:"applies_to_asset_types" json:"applies_to_asset_types,omitempty"`
	ImpactC             bool    `db:"impact_c"               json:"impact_c"`
	ImpactI             bool    `db:"impact_i"               json:"impact_i"`
	ImpactA             bool    `db:"impact_a"               json:"impact_a"`
	Status              *string `db:"status"                 json:"status,omitempty"`

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

// AssetVulnerability — конкретная БДУ-уязвимость, обнаруженная (или вручную
// добавленная) на активе. После P6 это инвентарь свидетельств наличия
// VL-категории на активе, а не FK на legacy-таблицу `vulnerabilities`.
type AssetVulnerability struct {
	ID            int64                    `db:"id"            json:"id"`
	AssetID       int64                    `db:"asset_id"      json:"asset_id"`
	BDUID         string                   `db:"bdu_id"        json:"bdu_id"`
	CVE           *string                  `db:"cve"           json:"cve,omitempty"`
	CWE           *string                  `db:"cwe"           json:"cwe,omitempty"`
	VLCategoryID  *int16                   `db:"vl_category_id" json:"vl_category_id,omitempty"`
	CVSSScore     *float64                 `db:"cvss_score"    json:"cvss_score,omitempty"`
	SeverityLevel *int16                   `db:"severity_level" json:"severity_level,omitempty"`
	Title         *string                  `db:"title"         json:"title,omitempty"`
	Source        string                   `db:"source"        json:"source"` // "auto:asset_software" | "manual"
	SoftwareID    *int64                   `db:"software_id"   json:"software_id,omitempty"`
	Status        AssetVulnerabilityStatus `db:"status"        json:"status"`
	DiscoveredAt  time.Time                `db:"discovered_at" json:"discovered_at"`
	CreatedAt     time.Time                `db:"created_at"    json:"created_at"`
	UpdatedAt     time.Time                `db:"updated_at"    json:"updated_at"`
}

// ---- Controls ----

type ControlType struct {
	ID          int16  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type Control struct {
	ID            int64     `db:"id"              json:"id"`
	Name          string    `db:"name"            json:"name"`
	ControlTypeID *int16    `db:"control_type_id" json:"control_type_id,omitempty"`
	Description   *string   `db:"description"     json:"description,omitempty"`
	CreatedAt     time.Time `db:"created_at"      json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"      json:"updated_at"`
}

type AssetControl struct {
	ID        int64     `db:"id"`
	AssetID   int64     `db:"asset_id"`
	ControlID int64     `db:"control_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

	IsRussian      bool    `db:"is_russian" json:"is_russian"`
	RegistryNumber *string `db:"registry_number" json:"registry_number,omitempty"`
	RegistryDate   *string `db:"registry_date" json:"registry_date,omitempty"`
	RegistryURL    *string `db:"registry_url" json:"registry_url,omitempty"`

	FSTECCertified       bool    `db:"fstec_certified" json:"fstec_certified"`
	FSTECCertificateNum  *string `db:"fstec_certificate_num" json:"fstec_certificate_num,omitempty"`
	FSTECCertificateDate *string `db:"fstec_certificate_date" json:"fstec_certificate_date,omitempty"`
	FSTECProtectionClass *string `db:"fstec_protection_class" json:"fstec_protection_class,omitempty"`
	FSTECValidUntil      *string `db:"fstec_valid_until" json:"fstec_valid_until,omitempty"`

	FSBCertified       bool    `db:"fsb_certified" json:"fsb_certified"`
	FSBCertificateNum  *string `db:"fsb_certificate_num" json:"fsb_certificate_num,omitempty"`
	FSBProtectionClass *string `db:"fsb_protection_class" json:"fsb_protection_class,omitempty"`

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

// SoftwareWithCategory — ПО с категорией для отображения.
type SoftwareWithCategory struct {
	Software
	CategoryName string `db:"category_name" json:"category_name,omitempty"`
}

// AssetSoftwareWithSoftware — запись о ПО на активе вместе с карточкой из справочника.
type AssetSoftwareWithSoftware struct {
	Link     AssetSoftware `json:"link"`
	Software Software      `json:"software"`
}

// ---- Compliance / ОСЗ (оценка состояния защищённости) ----

// ComplianceStandard — справочник стандартов ИБ (ФСТЭК-17, ISO 27001, …).
type ComplianceStandard struct {
	ID           int16   `db:"id"           json:"id"`
	Code         string  `db:"code"         json:"code"`
	Name         string  `db:"name"         json:"name"`
	FullName     string  `db:"full_name"    json:"full_name"`
	Jurisdiction string  `db:"jurisdiction" json:"jurisdiction"` // "RU" / "INT"
	Description  *string `db:"description"  json:"description,omitempty"`
	SortOrder    int16   `db:"sort_order"   json:"sort_order"`
}

// ComplianceRequirement — отдельное требование внутри стандарта.
type ComplianceRequirement struct {
	ID          int32   `db:"id"          json:"id"`
	StandardID  int16   `db:"standard_id" json:"standard_id"`
	Code        string  `db:"code"        json:"code"`
	Category    string  `db:"category"    json:"category"`
	Title       string  `db:"title"       json:"title"`
	Description *string `db:"description" json:"description,omitempty"`
	Priority    int16   `db:"priority"    json:"priority"`
	SortOrder   int16   `db:"sort_order"  json:"sort_order"`
}

// RequirementControlLink — ребро requirement ↔ control с весом покрытия.
type RequirementControlLink struct {
	RequirementID  int32   `db:"requirement_id"  json:"requirement_id"`
	ControlID      int64   `db:"control_id"      json:"control_id"`
	CoverageWeight float64 `db:"coverage_weight" json:"coverage_weight"`
}

// RequirementStatus — состояние одного требования для конкретного актива.
type RequirementStatus struct {
	Requirement      ComplianceRequirement `json:"requirement"`
	Coverage         float64               `json:"coverage"`            // [0..1]: max coverage_weight среди внедрённых control
	CoveringControls []Control             `json:"covering_controls"`   // что закрыло (внедрённые на активе)
	MissingControls  []Control             `json:"missing_controls"`    // что бы ещё закрыло, если внедрить
}

// AssetStandardCompliance — детализация по одному стандарту для одного актива.
type AssetStandardCompliance struct {
	Standard       ComplianceStandard  `json:"standard"`
	OverallScore   float64             `json:"overall_score"`   // [0..1]: avg(coverage) по всем требованиям
	CoveredCount   int                 `json:"covered_count"`   // # требований с coverage >= 1.0
	PartialCount   int                 `json:"partial_count"`   // 0 < coverage < 1.0
	UncoveredCount int                 `json:"uncovered_count"` // coverage == 0
	TotalCount     int                 `json:"total_count"`
	Requirements   []RequirementStatus `json:"requirements"`
}

// AssetComplianceOverview — короткая сводка по активу (для списка стандартов).
type AssetComplianceOverview struct {
	Standard       ComplianceStandard `json:"standard"`
	OverallScore   float64            `json:"overall_score"`
	CoveredCount   int                `json:"covered_count"`
	PartialCount   int                `json:"partial_count"`
	UncoveredCount int                `json:"uncovered_count"`
	TotalCount     int                `json:"total_count"`
}

// ---- Organization-level views ----

// OrganizationOverview — сводные показатели по всем активам организации.
type OrganizationOverview struct {
	TotalAssets         int                              `json:"total_assets"`
	IsolatedAssets      int                              `json:"isolated_assets"`
	AssetsByEnvironment map[string]int                   `json:"assets_by_environment"`
	AssetsByType        []AssetTypeBucket                `json:"assets_by_type"`
	RiskDistribution    map[string]int                   `json:"risk_distribution"`     // level -> кол-во активов
	WMax                float64                          `json:"w_max"`                 // максимальный W среди всех (asset×threat)
	WMaxAsset           string                           `json:"w_max_asset,omitempty"`
	WMaxThreat          string                           `json:"w_max_threat,omitempty"`
	AvgWPerAsset        float64                          `json:"avg_w_per_asset"`       // средний W_max по активам
	TotalControls       int                              `json:"total_controls"`        // суммарное количество (asset, control)
	UncoveredVLs        int                              `json:"uncovered_vls"`         // ∑ непокрытых VL по активам
	ComplianceByStd     []OrganizationComplianceSummary  `json:"compliance_by_standard"`
}

// AssetTypeBucket — распределение активов по типам.
type AssetTypeBucket struct {
	TypeID   *int16 `json:"type_id,omitempty"`
	TypeName string `json:"type_name"`
	Count    int    `json:"count"`
}

// OrganizationComplianceSummary — сводный compliance-score по стандарту
// усреднённый по всем активам.
type OrganizationComplianceSummary struct {
	Standard     ComplianceStandard `json:"standard"`
	AvgScore     float64            `json:"avg_score"`
	MinScore     float64            `json:"min_score"`
	MaxScore     float64            `json:"max_score"`
	AssetsCount  int                `json:"assets_count"`
}

// AssetMatrixRow — одна строка табличного представления актива в сводке организации.
type AssetMatrixRow struct {
	AssetID         int64                       `json:"asset_id"`
	Name            string                      `json:"name"`
	TypeName        string                      `json:"type_name,omitempty"`
	Environment     string                      `json:"environment,omitempty"`
	IsIsolated      bool                        `json:"is_isolated"`
	WMax            float64                     `json:"w_max"`
	Level           string                      `json:"level"`
	ThreatCount     int                         `json:"threat_count"`
	ControlCount    int                         `json:"control_count"`
	ComplianceByStd []AssetComplianceOverview   `json:"compliance_by_standard"`
}

// CriticalRisk — единичная (актив × угроза) пара с высоким W. Используется
// в «топ критичных рисков» сводного отчёта.
type CriticalRisk struct {
	AssetID    int64   `json:"asset_id"`
	AssetName  string  `json:"asset_name"`
	ThreatID   int64   `json:"threat_id"`
	ThreatName string  `json:"threat_name"`
	BDUID      string  `json:"bdu_id,omitempty"`
	W          float64 `json:"w"`
	Level      string  `json:"level"`
}
