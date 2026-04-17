package repository

import (
	"context"
	"encoding/json"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RiskGraphRepository interface {
	LoadVulnerableLinks(ctx context.Context, assetID, threatID int64) ([]domain.VLNode, error)
}

type riskGraphRepository struct {
	pool *pgxpool.Pool
}

func NewRiskGraphRepository(pool *pgxpool.Pool) RiskGraphRepository {
	return &riskGraphRepository{pool: pool}
}

func (r *riskGraphRepository) LoadVulnerableLinks(ctx context.Context, assetID, threatID int64) ([]domain.VLNode, error) {
	const q = `
WITH threat_vls AS (
    -- VL, относящиеся к угрозе
    SELECT v.id, v.name, v.severity
    FROM vulnerabilities v
    JOIN threat_vulnerable_links tvl ON tvl.vulnerability_id = v.id
    WHERE tvl.threat_id = $2
),
asset_vls AS (
    -- VL, которые реально присутствуют на активе
    SELECT v.id
    FROM vulnerabilities v
    JOIN asset_vulnerabilities av ON av.vulnerability_id = v.id
    WHERE av.asset_id = $1 AND av.status IN ('open','mitigated')
),
vl_covering_controls AS (
    -- контроли, закрывающие VL угрозы, И внедрённые на активе
    SELECT vc.vulnerability_id,
           c.id   AS control_id,
           c.name AS control_name,
           vc.coverage
    FROM vulnerability_controls vc
    JOIN controls c        ON c.id = vc.control_id
    JOIN asset_controls ac ON ac.control_id = c.id AND ac.asset_id = $1
)
SELECT tv.id,
       tv.name,
       tv.severity,
       COALESCE(
         json_agg(json_build_object(
           'id', vcc.control_id,
           'name', vcc.control_name,
           'coverage', vcc.coverage
         )) FILTER (WHERE vcc.control_id IS NOT NULL),
         '[]'::json
       ) AS controls_json,
       -- uncovered: присутствует на активе И не имеет ни одного covering control
       (EXISTS (SELECT 1 FROM asset_vls av WHERE av.id = tv.id))
         AND NOT EXISTS (
           SELECT 1 FROM vl_covering_controls vcc2 WHERE vcc2.vulnerability_id = tv.id
         ) AS uncovered
FROM threat_vls tv
LEFT JOIN vl_covering_controls vcc ON vcc.vulnerability_id = tv.id
GROUP BY tv.id, tv.name, tv.severity
ORDER BY tv.id`

	rows, err := r.pool.Query(ctx, q, assetID, threatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.VLNode, 0)
	for rows.Next() {
		var v domain.VLNode
		var raw []byte
		if err := rows.Scan(&v.VulnerabilityID, &v.Name, &v.Severity, &raw, &v.Uncovered); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &v.CoverageControls); err != nil {
				return nil, err
			}
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
