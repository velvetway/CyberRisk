package repository

import (
	"context"
	"fmt"
	"strings"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PTSZIRepository interface {
	ListThreats(ctx context.Context) ([]domain.PTSZIThreat, error)
	GetThreat(ctx context.Context, id int64) (*domain.PTSZIThreat, error)
	ListSources(ctx context.Context) ([]domain.ThreatSource, error)
	ListVulnerableLinks(ctx context.Context) ([]domain.PTSZIVulnerableLink, error)
	ListControls(ctx context.Context) ([]domain.PTSZIControl, error)
	ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error)
	AssetContour(ctx context.Context, assetID int64) (string, error)
	ListAssetVulnerableLinks(ctx context.Context, assetID int64) ([]domain.PTSZIAssetVulnerableLink, error)
	SaveAssetVulnerableLinks(ctx context.Context, assetID int64, ids []int16) error
	ListAssetControls(ctx context.Context, assetID int64) ([]domain.PTSZIAssetControl, error)
	SaveAssetControls(ctx context.Context, assetID int64, controls []AssetPTSZIControlInput) error
	LoadAttackPath(ctx context.Context, asset *domain.Asset, threatID int64) (*domain.PTSZIAttackPath, error)
	ListUBI(ctx context.Context, limit, offset int32, query string) ([]domain.PTSZIUBIThreat, error)
}

type AssetPTSZIControlInput struct {
	ControlID     int16
	Effectiveness float64
}

type ptsziRepository struct {
	pool *pgxpool.Pool
}

func NewPTSZIRepository(pool *pgxpool.Pool) PTSZIRepository {
	return &ptsziRepository{pool: pool}
}

func (r *ptsziRepository) ListThreats(ctx context.Context) ([]domain.PTSZIThreat, error) {
	rows, err := r.pool.Query(ctx, `
SELECT t.id, t.code, t.name, t.description, t.q_threat, t.q_severity,
       COALESCE(string_agg(tc.contour, ',' ORDER BY tc.contour), '') AS contours
FROM ptszi_threats t
LEFT JOIN ptszi_threat_contours tc ON tc.threat_id = t.id
GROUP BY t.id, t.code, t.name, t.description, t.q_threat, t.q_severity
ORDER BY t.code`)
	if err != nil {
		return nil, fmt.Errorf("list ptszi threats: %w", err)
	}
	defer rows.Close()
	return scanPTSZIThreats(rows)
}

func (r *ptsziRepository) GetThreat(ctx context.Context, id int64) (*domain.PTSZIThreat, error) {
	rows, err := r.pool.Query(ctx, `
SELECT t.id, t.code, t.name, t.description, t.q_threat, t.q_severity,
       COALESCE(string_agg(tc.contour, ',' ORDER BY tc.contour), '') AS contours
FROM ptszi_threats t
LEFT JOIN ptszi_threat_contours tc ON tc.threat_id = t.id
WHERE t.id = $1
GROUP BY t.id, t.code, t.name, t.description, t.q_threat, t.q_severity`, id)
	if err != nil {
		return nil, fmt.Errorf("get ptszi threat: %w", err)
	}
	defer rows.Close()
	items, err := scanPTSZIThreats(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func scanPTSZIThreats(rows pgx.Rows) ([]domain.PTSZIThreat, error) {
	out := make([]domain.PTSZIThreat, 0)
	for rows.Next() {
		var t domain.PTSZIThreat
		var desc *string
		var contours string
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &desc, &t.QThreat, &t.QSeverity, &contours); err != nil {
			return nil, fmt.Errorf("scan ptszi threat: %w", err)
		}
		t.Description = desc
		t.Contours = splitCSV(contours)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ListSources(ctx context.Context) ([]domain.ThreatSource, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, COALESCE(description, ''), created_at FROM threat_sources ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("list threat sources: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ThreatSource, 0)
	for rows.Next() {
		var s domain.ThreatSource
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan threat source: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ListVulnerableLinks(ctx context.Context) ([]domain.PTSZIVulnerableLink, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, description FROM ptszi_vulnerable_links ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("list ptszi vulnerable links: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIVulnerableLink, 0)
	for rows.Next() {
		var v domain.PTSZIVulnerableLink
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Description); err != nil {
			return nil, fmt.Errorf("scan ptszi vulnerable link: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ListControls(ctx context.Context) ([]domain.PTSZIControl, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, description FROM ptszi_controls ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("list ptszi controls: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIControl, 0)
	for rows.Next() {
		var c domain.PTSZIControl
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description); err != nil {
			return nil, fmt.Errorf("scan ptszi control: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, code, name, affects_confidentiality, affects_integrity, affects_availability,
       COALESCE(description, ''), created_at
FROM destructive_actions
ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("list destructive actions: %w", err)
	}
	defer rows.Close()
	out := make([]domain.DestructiveAction, 0)
	for rows.Next() {
		var da domain.DestructiveAction
		if err := rows.Scan(&da.ID, &da.Code, &da.Name, &da.AffectsConfidentiality, &da.AffectsIntegrity, &da.AffectsAvailability, &da.Description, &da.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan destructive action: %w", err)
		}
		out = append(out, da)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ListAssetVulnerableLinks(ctx context.Context, assetID int64) ([]domain.PTSZIAssetVulnerableLink, error) {
	rows, err := r.pool.Query(ctx, `
SELECT vl.id, vl.code, vl.name, vl.description, avl.status, avl.comment
FROM asset_vulnerable_links avl
JOIN ptszi_vulnerable_links vl ON vl.id = avl.vulnerable_link_id
WHERE avl.asset_id = $1
ORDER BY vl.code`, assetID)
	if err != nil {
		return nil, fmt.Errorf("list asset ptszi vulnerable links: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIAssetVulnerableLink, 0)
	for rows.Next() {
		var item domain.PTSZIAssetVulnerableLink
		if err := rows.Scan(&item.VulnerableLink.ID, &item.VulnerableLink.Code, &item.VulnerableLink.Name, &item.VulnerableLink.Description, &item.Status, &item.Comment); err != nil {
			return nil, fmt.Errorf("scan asset ptszi vulnerable link: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) SaveAssetVulnerableLinks(ctx context.Context, assetID int64, ids []int16) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM asset_vulnerable_links WHERE asset_id = $1`, assetID); err != nil {
		return fmt.Errorf("delete asset ptszi vulnerable links: %w", err)
	}
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `
INSERT INTO asset_vulnerable_links (asset_id, vulnerable_link_id, status)
VALUES ($1, $2, 'present')
ON CONFLICT (asset_id, vulnerable_link_id) DO UPDATE SET status = 'present', updated_at = now()`, assetID, id); err != nil {
			return fmt.Errorf("insert asset ptszi vulnerable link: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *ptsziRepository) ListAssetControls(ctx context.Context, assetID int64) ([]domain.PTSZIAssetControl, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.code, c.name, c.description, ac.effectiveness, ac.comment
FROM asset_ptszi_controls ac
JOIN ptszi_controls c ON c.id = ac.control_id
WHERE ac.asset_id = $1
ORDER BY c.code`, assetID)
	if err != nil {
		return nil, fmt.Errorf("list asset ptszi controls: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIAssetControl, 0)
	for rows.Next() {
		var item domain.PTSZIAssetControl
		if err := rows.Scan(&item.Control.ID, &item.Control.Code, &item.Control.Name, &item.Control.Description, &item.Effectiveness, &item.Comment); err != nil {
			return nil, fmt.Errorf("scan asset ptszi control: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) SaveAssetControls(ctx context.Context, assetID int64, controls []AssetPTSZIControlInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM asset_ptszi_controls WHERE asset_id = $1`, assetID); err != nil {
		return fmt.Errorf("delete asset ptszi controls: %w", err)
	}
	for _, c := range controls {
		eff := c.Effectiveness
		if eff < 0 {
			eff = 0
		}
		if eff > 1 {
			eff = 1
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO asset_ptszi_controls (asset_id, control_id, effectiveness)
VALUES ($1, $2, $3)
ON CONFLICT (asset_id, control_id) DO UPDATE SET effectiveness = EXCLUDED.effectiveness, updated_at = now()`, assetID, c.ControlID, eff); err != nil {
			return fmt.Errorf("insert asset ptszi control: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *ptsziRepository) LoadAttackPath(ctx context.Context, asset *domain.Asset, threatID int64) (*domain.PTSZIAttackPath, error) {
	threat, err := r.GetThreat(ctx, threatID)
	if err != nil {
		return nil, err
	}
	if threat == nil {
		return nil, nil
	}
	assetContour, err := r.AssetContour(ctx, asset.ID)
	if err != nil {
		return nil, err
	}
	path := &domain.PTSZIAttackPath{
		Asset:        domain.AssetRef{ID: asset.ID, Name: asset.Name},
		AssetContour: assetContour,
		Threat:       *threat,
		QThreat:      threat.QThreat,
		QSeverity:    threat.QSeverity,
		Z:            zFromContours(threat.Contours),
		Applicable:   containsString(threat.Contours, assetContour),
	}
	path.Sources, err = r.sourcesForThreat(ctx, threatID)
	if err != nil {
		return nil, err
	}
	path.DestructiveActions, err = r.dasForThreat(ctx, threatID)
	if err != nil {
		return nil, err
	}
	path.VulnerableLinks, err = r.actualVLsForThreat(ctx, asset.ID, threatID)
	if err != nil {
		return nil, err
	}
	path.UBI, err = r.ubiForThreat(ctx, threatID, 10)
	if err != nil {
		return nil, err
	}
	path.Applicable = path.Applicable && len(path.VulnerableLinks) > 0
	return path, nil
}

func (r *ptsziRepository) AssetContour(ctx context.Context, assetID int64) (string, error) {
	var contour string
	if err := r.pool.QueryRow(ctx, `SELECT security_contour FROM assets WHERE id = $1`, assetID).Scan(&contour); err != nil {
		return "", fmt.Errorf("get asset contour: %w", err)
	}
	return contour, nil
}

func (r *ptsziRepository) sourcesForThreat(ctx context.Context, threatID int64) ([]domain.ThreatSource, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id, s.code, s.name, COALESCE(s.description, ''), s.created_at
FROM ptszi_source_threats st
JOIN threat_sources s ON s.id = st.source_id
WHERE st.threat_id = $1
ORDER BY s.code`, threatID)
	if err != nil {
		return nil, fmt.Errorf("load ptszi sources: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ThreatSource, 0)
	for rows.Next() {
		var s domain.ThreatSource
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ptszi source: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) dasForThreat(ctx context.Context, threatID int64) ([]domain.DestructiveAction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT da.id, da.code, da.name, da.affects_confidentiality, da.affects_integrity,
       da.affects_availability, COALESCE(da.description, ''), da.created_at
FROM ptszi_threat_destructive_actions tda
JOIN destructive_actions da ON da.id = tda.destructive_action_id
WHERE tda.threat_id = $1
ORDER BY da.code`, threatID)
	if err != nil {
		return nil, fmt.Errorf("load ptszi destructive actions: %w", err)
	}
	defer rows.Close()
	out := make([]domain.DestructiveAction, 0)
	for rows.Next() {
		var da domain.DestructiveAction
		if err := rows.Scan(&da.ID, &da.Code, &da.Name, &da.AffectsConfidentiality, &da.AffectsIntegrity, &da.AffectsAvailability, &da.Description, &da.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ptszi destructive action: %w", err)
		}
		out = append(out, da)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) actualVLsForThreat(ctx context.Context, assetID, threatID int64) ([]domain.PTSZIPathVL, error) {
	rows, err := r.pool.Query(ctx, `
SELECT vl.id, vl.code, vl.name, vl.description, avl.status, avl.comment
FROM ptszi_threat_vulnerable_links tvl
JOIN asset_vulnerable_links avl ON avl.vulnerable_link_id = tvl.vulnerable_link_id
JOIN ptszi_vulnerable_links vl ON vl.id = tvl.vulnerable_link_id
WHERE tvl.threat_id = $2 AND avl.asset_id = $1 AND avl.status = 'present'
ORDER BY vl.code`, assetID, threatID)
	if err != nil {
		return nil, fmt.Errorf("load actual ptszi vls: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIPathVL, 0)
	for rows.Next() {
		var v domain.PTSZIPathVL
		if err := rows.Scan(&v.VulnerableLink.ID, &v.VulnerableLink.Code, &v.VulnerableLink.Name, &v.VulnerableLink.Description, &v.Status, &v.Comment); err != nil {
			return nil, fmt.Errorf("scan actual ptszi vl: %w", err)
		}
		v.Controls, err = r.controlsForVL(ctx, assetID, v.VulnerableLink.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) controlsForVL(ctx context.Context, assetID int64, vlID int16) ([]domain.PTSZIControlCoverage, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.code, c.name, c.description, vlc.coverage,
       COALESCE(ac.effectiveness, 0) AS effectiveness,
       ac.control_id IS NOT NULL AS implemented
FROM ptszi_vulnerable_link_controls vlc
JOIN ptszi_controls c ON c.id = vlc.control_id
LEFT JOIN asset_ptszi_controls ac ON ac.control_id = c.id AND ac.asset_id = $1
WHERE vlc.vulnerable_link_id = $2
ORDER BY c.code`, assetID, vlID)
	if err != nil {
		return nil, fmt.Errorf("load ptszi vl controls: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PTSZIControlCoverage, 0)
	for rows.Next() {
		var c domain.PTSZIControlCoverage
		if err := rows.Scan(&c.Control.ID, &c.Control.Code, &c.Control.Name, &c.Control.Description, &c.Coverage, &c.Effectiveness, &c.Implemented); err != nil {
			return nil, fmt.Errorf("scan ptszi vl control: %w", err)
		}
		c.ResultingCoverage = c.Coverage * c.Effectiveness
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ptsziRepository) ubiForThreat(ctx context.Context, threatID int64, limit int) ([]domain.PTSZIUBIThreat, error) {
	rows, err := r.pool.Query(ctx, `
SELECT u.id, u.ubi_code, u.ubi_number, u.name, u.description, u.source_raw, u.impact_object,
       u.impact_confidentiality, u.impact_integrity, u.impact_availability,
       u.max_potential, u.q_threat, u.q_severity,
       COALESCE(string_agg(DISTINCT s.code, ',' ORDER BY s.code), '') AS mapped_sources
FROM ptszi_threat_ubi_links tul
JOIN ptszi_ubi_threats u ON u.ubi_code = tul.ubi_code
LEFT JOIN ptszi_ubi_source_mappings sm ON sm.ubi_code = u.ubi_code
LEFT JOIN threat_sources s ON s.id = sm.source_id
WHERE tul.threat_id = $1
GROUP BY u.id, u.ubi_code, u.ubi_number, u.name, u.description, u.source_raw, u.impact_object,
         u.impact_confidentiality, u.impact_integrity, u.impact_availability,
         u.max_potential, u.q_threat, u.q_severity
ORDER BY u.ubi_number
LIMIT $2`, threatID, limit)
	if err != nil {
		return nil, fmt.Errorf("load ptszi ubi for threat: %w", err)
	}
	defer rows.Close()
	return scanUBIThreats(rows)
}

func (r *ptsziRepository) ListUBI(ctx context.Context, limit, offset int32, query string) ([]domain.PTSZIUBIThreat, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{}
	q := `
SELECT u.id, u.ubi_code, u.ubi_number, u.name, u.description, u.source_raw, u.impact_object,
       u.impact_confidentiality, u.impact_integrity, u.impact_availability,
       u.max_potential, u.q_threat, u.q_severity,
       COALESCE(string_agg(DISTINCT s.code, ',' ORDER BY s.code), '') AS mapped_sources
FROM ptszi_ubi_threats u
LEFT JOIN ptszi_ubi_source_mappings sm ON sm.ubi_code = u.ubi_code
LEFT JOIN threat_sources s ON s.id = sm.source_id
WHERE 1=1`
	if query != "" {
		args = append(args, "%"+query+"%")
		q += fmt.Sprintf(" AND (u.ubi_code ILIKE $%d OR u.name ILIKE $%d OR u.impact_object ILIKE $%d)", len(args), len(args), len(args))
	}
	args = append(args, limit, offset)
	q += fmt.Sprintf(`
GROUP BY u.id, u.ubi_code, u.ubi_number, u.name, u.description, u.source_raw, u.impact_object,
         u.impact_confidentiality, u.impact_integrity, u.impact_availability,
         u.max_potential, u.q_threat, u.q_severity
ORDER BY u.ubi_number
LIMIT $%d OFFSET $%d`, len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list ptszi ubi: %w", err)
	}
	defer rows.Close()
	return scanUBIThreats(rows)
}

func scanUBIThreats(rows pgx.Rows) ([]domain.PTSZIUBIThreat, error) {
	out := make([]domain.PTSZIUBIThreat, 0)
	for rows.Next() {
		var u domain.PTSZIUBIThreat
		var sources string
		if err := rows.Scan(
			&u.ID, &u.UBICode, &u.UBINumber, &u.Name, &u.Description, &u.SourceRaw, &u.ImpactObject,
			&u.ImpactConfidentiality, &u.ImpactIntegrity, &u.ImpactAvailability,
			&u.MaxPotential, &u.QThreat, &u.QSeverity, &sources,
		); err != nil {
			return nil, fmt.Errorf("scan ptszi ubi: %w", err)
		}
		u.MappedSources = splitCSV(sources)
		out = append(out, u)
	}
	return out, rows.Err()
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if x := strings.TrimSpace(p); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func zFromContours(contours []string) float64 {
	if containsString(contours, "external") && containsString(contours, "internal") {
		return 1.0
	}
	return 0.5
}
