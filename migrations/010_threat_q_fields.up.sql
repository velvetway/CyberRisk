-- Add Q-threat and q-severity fields for W_i formula
ALTER TABLE threats
    ADD COLUMN IF NOT EXISTS q_threat   NUMERIC(3,2) NOT NULL DEFAULT 0.5
        CHECK (q_threat BETWEEN 0 AND 1),
    ADD COLUMN IF NOT EXISTS q_severity NUMERIC(3,2) NOT NULL DEFAULT 0.5
        CHECK (q_severity BETWEEN 0 AND 1);

-- Backfill: q_threat from base_likelihood (1..5 → 0.1..0.9)
UPDATE threats SET q_threat = GREATEST(0.1, LEAST(0.9, base_likelihood::NUMERIC / 5.0));

-- Backfill: q_severity from CIA impact flags
--   sum of flags / 3, clipped to [0.2, 1.0]
UPDATE threats SET q_severity = GREATEST(0.2, LEAST(1.0,
    (COALESCE(impact_confidentiality::INT, 0)
   + COALESCE(impact_integrity::INT, 0)
   + COALESCE(impact_availability::INT, 0))::NUMERIC / 3.0
));

-- Seed edges: S -> ST (канонические связи из модели ПТСЗИ)
INSERT INTO source_threats (threat_source_id, threat_id)
SELECT s.id, t.id
FROM threat_sources s
CROSS JOIN threats t
WHERE
  (s.code='S3' AND (t.name ILIKE '%вирус%' OR t.name ILIKE '%malware%'))
  OR
  (s.code='S4' AND (t.name ILIKE '%вирус%' OR t.name ILIKE '%несанкциониров%'
                 OR t.name ILIKE '%DDoS%' OR t.name ILIKE '%перехват%'
                 OR t.name ILIKE '%проникнов%' OR t.name ILIKE '%сканиров%'
                 OR t.name ILIKE '%spam%' OR t.name ILIKE '%подмен%'))
  OR
  (s.code='S1' AND (t.name ILIKE '%несанкциониров%' OR t.name ILIKE '%DDoS%'
                 OR t.name ILIKE '%перехват%' OR t.name ILIKE '%сканиров%'
                 OR t.name ILIKE '%spam%' OR t.name ILIKE '%подмен%'))
  OR
  (s.code='S2' AND (t.name ILIKE '%несанкциониров%' OR t.name ILIKE '%spam%' OR t.name ILIKE '%подмен%'))
ON CONFLICT DO NOTHING;

-- Seed edges: ST -> DA
INSERT INTO threat_destructive_actions (threat_id, destructive_action_id)
SELECT t.id, da.id
FROM threats t
CROSS JOIN destructive_actions da
WHERE
  (t.impact_confidentiality AND da.code IN ('DA1','DA2','DA6'))
  OR
  (t.impact_integrity       AND da.code IN ('DA4','DA7'))
  OR
  (t.impact_availability    AND da.code IN ('DA3','DA5','DA7'))
ON CONFLICT DO NOTHING;

-- Seed edges: ST -> VL (heuristic by name)
INSERT INTO threat_vulnerable_links (threat_id, vulnerability_id)
SELECT t.id, v.id
FROM threats t, vulnerabilities v
WHERE
  (t.name ILIKE '%brute%' AND (v.name ILIKE '%пароль%' OR v.name ILIKE '%password%'))
  OR
  (t.name ILIKE '%вирус%' AND (v.name ILIKE '%устарев%' OR v.name ILIKE '%outdated%'))
  OR
  (t.name ILIKE '%перехват%' AND (v.name ILIKE '%шифров%' OR v.name ILIKE '%encryption%'))
  OR
  (t.name ILIKE '%DDoS%' AND (v.name ILIKE '%фильтр%' OR v.name ILIKE '%rate%'))
ON CONFLICT DO NOTHING;

-- Seed VL -> Control: heuristic by name keywords
INSERT INTO vulnerability_controls (vulnerability_id, control_id, coverage)
SELECT v.id, c.id, 1.0
FROM vulnerabilities v, controls c
WHERE
  (v.name ILIKE '%пароль%'  AND c.name ILIKE '%MFA%')
  OR
  (v.name ILIKE '%устарев%' AND c.name ILIKE '%patch%')
  OR
  (v.name ILIKE '%шифров%'  AND c.name ILIKE '%encrypt%')
  OR
  (v.name ILIKE '%фильтр%'  AND c.name ILIKE '%firewall%')
ON CONFLICT DO NOTHING;
