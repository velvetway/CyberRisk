DROP INDEX IF EXISTS idx_ptszi_threat_contours_contour;
DROP INDEX IF EXISTS idx_asset_ptszi_controls_asset;
DROP INDEX IF EXISTS idx_asset_vulnerable_links_asset;

DROP TABLE IF EXISTS ptszi_recommendation_templates;
DROP TABLE IF EXISTS ptszi_threat_ubi_links;
DROP TABLE IF EXISTS ptszi_source_threats;
DROP TABLE IF EXISTS ptszi_threat_destructive_actions;
DROP TABLE IF EXISTS ptszi_vulnerable_link_controls;
DROP TABLE IF EXISTS ptszi_threat_vulnerable_links;
DROP TABLE IF EXISTS asset_ptszi_controls;
DROP TABLE IF EXISTS ptszi_controls;
DROP TABLE IF EXISTS asset_vulnerable_links;
DROP TABLE IF EXISTS ptszi_vulnerable_links;
DROP TABLE IF EXISTS ptszi_threat_contours;
DROP TABLE IF EXISTS ptszi_threats;

ALTER TABLE assets
    DROP COLUMN IF EXISTS document_flow,
    DROP COLUMN IF EXISTS existing_security_tools,
    DROP COLUMN IF EXISTS communication_channels,
    DROP COLUMN IF EXISTS has_remote_access,
    DROP COLUMN IF EXISTS users_description,
    DROP COLUMN IF EXISTS processed_information,
    DROP COLUMN IF EXISTS network_location,
    DROP COLUMN IF EXISTS system_composition,
    DROP COLUMN IF EXISTS purpose,
    DROP COLUMN IF EXISTS security_contour;
