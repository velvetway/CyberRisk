-- Wipe the legacy demo content of threats / vulnerabilities / controls so the
-- ФСТЭК / БДУ / Минцифры importers (P2..P4 of the FSTEC integration plan) can
-- repopulate from authoritative sources.
--
-- CASCADE clears every dependent edge in one shot:
--   threats              → threat_vulnerable_links, threat_destructive_actions,
--                          source_threats
--   vulnerabilities      → asset_vulnerabilities, vulnerability_controls
--   controls             → asset_controls (+ remaining vulnerability_controls)
--
-- Reference catalogues (asset_types, threat_categories, vulnerability_categories,
-- control_types, threat_sources, destructive_actions) are deliberately kept —
-- they will be reused by the new data.
--
-- User-created assets are kept; their orphaned VL/control links die with the
-- CASCADE above.

TRUNCATE TABLE threats, vulnerabilities, controls RESTART IDENTITY CASCADE;
