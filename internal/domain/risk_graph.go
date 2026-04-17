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

// VLNode — уязвимое звено плюс средства защиты, закрывающие его на конкретном активе.
type VLNode struct {
	VulnerabilityID  int64             `json:"vulnerability_id"`
	Name             string            `json:"name"`
	Severity         int               `json:"severity"`           // 1..10
	CoverageControls []ControlCoverage `json:"coverage_controls"`  // controls present on this asset that cover this VL
	Uncovered        bool              `json:"uncovered"`
}

// ControlCoverage is the runtime view of a control that covers a given VL.
// (Separate from the `Control` DB struct if one exists elsewhere; this type is for graph output.)
type ControlCoverage struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Coverage float64 `json:"coverage"` // 0..1 from vulnerability_controls.coverage
}
