-- Wipe the hand-crafted software_catalog seed (≈30 rows) before the Минцифры
-- importer (P4) replaces it with the full реестр (~26 094 entries).
--
-- asset_software FK is ON DELETE CASCADE → user-attached pieces of software
-- on assets will be re-detached cleanly. The user can re-attach via the new
-- detail UI in P8.

TRUNCATE TABLE software_catalog RESTART IDENTITY CASCADE;
