-- Reverting to VARCHAR(256) would fail on rows whose name is longer; we
-- intentionally keep the wider TEXT type even on rollback.
SELECT 1;
