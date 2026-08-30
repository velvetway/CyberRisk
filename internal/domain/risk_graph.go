package domain

// AttackPath — развёрнутая цепочка S → ST → VL → DA для одного актива и одной угрозы.
type AttackPath struct {
	Asset              AssetRef            `json:"asset"`
	Threat             ThreatRef           `json:"threat"`
	Sources            []ThreatSource      `json:"sources"`
	VulnerableLinks    []VLNode            `json:"vulnerable_links"`
	DestructiveActions []DestructiveAction `json:"destructive_actions"`
	W                  float64             `json:"w"` // [0,1]
	QThreat            float64             `json:"q_threat"`
	QSeverity          float64             `json:"q_severity"`
	QReaction          float64             `json:"q_reaction"`
	Z                  float64             `json:"z"`
	Level              string              `json:"level"` // low/medium/high/critical
}

type AssetRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ThreatRef struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	BDUID string `json:"bdu_id,omitempty"`
}

// VLNode — VL-категория из диплома (всего 6 шт), плюс контроли,
// закрывающие её на конкретном активе, и presence-индикатор —
// сколько CVE/БДУ-записей соответствующей категории сейчас активны
// в инвентаре актива (asset_vulnerabilities). Используется UI-картой
// «найдено N свидетельств», но не меняет формулу W в P6.
type VLNode struct {
	CategoryID       int16             `json:"category_id"`
	Code             string            `json:"code"` // VL1..VL6
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	CoverageControls []ControlCoverage `json:"coverage_controls"`
	Uncovered        bool              `json:"uncovered"`
	PresenceCount    int               `json:"presence_count"`
}

// ControlCoverage is the runtime view of a control that covers a given VL.
// (Separate from the `Control` DB struct if one exists elsewhere; this type is for graph output.)
type ControlCoverage struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Coverage float64 `json:"coverage"` // 0..1 from vl_category_controls.coverage
}

// AssetAggregate — сводные метрики по всем угрозам одного актива.
type AssetAggregate struct {
	WMax           float64 `json:"w_max"`
	Level          string  `json:"level"`
	ThreatCount    int     `json:"threat_count"`
	UncoveredCount int     `json:"uncovered_count"`
}

// AssetAttackPathsResponse — ответ bulk-эндпоинта /api/risk/asset/:asset_id/attack-paths.
type AssetAttackPathsResponse struct {
	Asset     AssetRef       `json:"asset"`
	Aggregate AssetAggregate `json:"aggregate"`
	Paths     []AttackPath   `json:"paths"`
}
