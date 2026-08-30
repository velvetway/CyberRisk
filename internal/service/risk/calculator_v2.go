package risk

import "Diplom/internal/domain"

// CalculateW implements the PTSZI formula
//
//	W_i = (Q_threat + q_severity + (1 - Q_reaction)) / 3 * Z
//
// All Q inputs are clamped to [0,1]; Z is clamped to {0.5, 1.0}
// (any input ≥ 1.0 → 1.0, otherwise 0.5).
func CalculateW(qThreat, qSeverity, qReaction, z float64) float64 {
	qThreat = clamp01(qThreat)
	qSeverity = clamp01(qSeverity)
	qReaction = clamp01(qReaction)
	if z >= 1.0 {
		z = 1.0
	} else {
		z = 0.5
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

// QReactionFromVLs is the share of a threat's vulnerable links that have at least
// one control with non-zero coverage deployed on the asset.
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

// ZFromAsset returns the contour-criticality coefficient strictly per the PTSZI thesis:
//
//	is_isolated = TRUE  → Z = 0.5  (актив виден только в одном контуре)
//	is_isolated = FALSE → Z = 1.0  (актив актуален для обоих контуров)
func ZFromAsset(a domain.Asset) float64 {
	if a.IsIsolated {
		return 0.5
	}
	return 1.0
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
