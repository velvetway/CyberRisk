package risk

import "Diplom/internal/domain"

// CalculateW implements the PTSZI formula W = (Q_threat + q_severity + (1 - Q_reaction)) / 3 * Z.
// All Q inputs are clamped to [0,1]; Z is clamped to [0.5, 1.0].
func CalculateW(qThreat, qSeverity, qReaction, z float64) float64 {
	qThreat = clamp01(qThreat)
	qSeverity = clamp01(qSeverity)
	qReaction = clamp01(qReaction)
	if z < 0.5 {
		z = 0.5
	} else if z > 1.0 {
		z = 1.0
	}
	return (qThreat + qSeverity + (1.0 - qReaction)) / 3.0 * z
}

// LevelFromW maps a W score to a qualitative risk level.
func LevelFromW(w float64) string {
	switch {
	case w >= 0.75:
		return "critical"
	case w >= 0.50:
		return "high"
	case w >= 0.25:
		return "medium"
	default:
		return "low"
	}
}

// QReactionFromVLs is the share of a threat's vulnerable links that have at least one
// control with non-zero coverage deployed on the asset.
// Returns 0 when the threat has no VLs (i.e., nothing can be "covered").
func QReactionFromVLs(vls []domain.VLNode) float64 {
	if len(vls) == 0 {
		return 0.0
	}
	covered := 0
	for _, v := range vls {
		for _, c := range v.CoverageControls {
			if c.Coverage > 0 {
				covered++
				break
			}
		}
	}
	return float64(covered) / float64(len(vls))
}

// ZFromAsset assigns the contour-criticality coefficient to an asset:
//
//	isolated        → 0.5  (угроза актуальна только для одного контура)
//	prod            → 1.0
//	stage/staging   → 0.75
//	otherwise (dev) → 0.5
func ZFromAsset(a domain.Asset) float64 {
	if a.IsIsolated {
		return 0.5
	}
	switch a.Environment {
	case "prod", "production":
		return 1.0
	case "stage", "staging":
		return 0.75
	default:
		return 0.5
	}
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}
