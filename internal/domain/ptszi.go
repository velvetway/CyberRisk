package domain

type PTSZIThreat struct {
	ID          int64    `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	QThreat     float64  `json:"q_threat"`
	QSeverity   float64  `json:"q_severity"`
	Contours    []string `json:"contours,omitempty"`
}

type PTSZIVulnerableLink struct {
	ID          int16   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type PTSZIControl struct {
	ID          int16   `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type PTSZIAssetVulnerableLink struct {
	VulnerableLink PTSZIVulnerableLink `json:"vulnerable_link"`
	Status         string              `json:"status"`
	Comment        *string             `json:"comment,omitempty"`
}

type PTSZIAssetControl struct {
	Control       PTSZIControl `json:"control"`
	Effectiveness float64      `json:"effectiveness"`
	Comment       *string      `json:"comment,omitempty"`
}

type PTSZIUBIThreat struct {
	ID                    int64    `json:"id"`
	UBICode               string   `json:"ubi_code"`
	UBINumber             int      `json:"ubi_number"`
	Name                  string   `json:"name"`
	Description           *string  `json:"description,omitempty"`
	SourceRaw             *string  `json:"source_raw,omitempty"`
	ImpactObject          *string  `json:"impact_object,omitempty"`
	ImpactConfidentiality bool     `json:"impact_confidentiality"`
	ImpactIntegrity       bool     `json:"impact_integrity"`
	ImpactAvailability    bool     `json:"impact_availability"`
	MaxPotential          string   `json:"max_potential"`
	QThreat               float64  `json:"q_threat"`
	QSeverity             float64  `json:"q_severity"`
	MappedSources         []string `json:"mapped_sources,omitempty"`
}

type PTSZIAttackPath struct {
	Asset              AssetRef              `json:"asset"`
	AssetContour       string                `json:"asset_contour"`
	Threat             PTSZIThreat           `json:"threat"`
	Sources            []ThreatSource        `json:"sources"`
	VulnerableLinks    []PTSZIPathVL         `json:"vulnerable_links"`
	DestructiveActions []DestructiveAction   `json:"destructive_actions"`
	UBI                []PTSZIUBIThreat      `json:"ubi"`
	QThreat            float64               `json:"q_threat"`
	QSeverity          float64               `json:"q_severity"`
	QReaction          float64               `json:"q_reaction"`
	Z                  float64               `json:"z"`
	W                  float64               `json:"w"`
	Level              string                `json:"level"`
	Applicable         bool                  `json:"applicable"`
	MissingControls    []PTSZIControl        `json:"missing_controls"`
	Recommendations    []PTSZIRecommendation `json:"recommendations,omitempty"`
}

type PTSZIPathVL struct {
	VulnerableLink PTSZIVulnerableLink    `json:"vulnerable_link"`
	Status         string                 `json:"status"`
	Comment        *string                `json:"comment,omitempty"`
	Coverage       float64                `json:"coverage"`
	Uncovered      bool                   `json:"uncovered"`
	Controls       []PTSZIControlCoverage `json:"controls"`
}

type PTSZIControlCoverage struct {
	Control           PTSZIControl `json:"control"`
	Coverage          float64      `json:"coverage"`
	Implemented       bool         `json:"implemented"`
	Effectiveness     float64      `json:"effectiveness"`
	ResultingCoverage float64      `json:"resulting_coverage"`
}

type PTSZIRecommendation struct {
	ControlID   int16  `json:"control_id"`
	ControlCode string `json:"control_code"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type PTSZIAssetProfile struct {
	Asset             Asset                      `json:"asset"`
	SecurityContour   string                     `json:"security_contour"`
	VulnerableLinks   []PTSZIAssetVulnerableLink `json:"vulnerable_links"`
	Controls          []PTSZIAssetControl        `json:"controls"`
	ApplicableThreats []PTSZIAttackPath          `json:"applicable_threats"`
}
