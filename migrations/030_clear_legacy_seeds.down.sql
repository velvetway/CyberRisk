-- Intentional no-op: the legacy demo rows are gone for good. Replaying every
-- earlier migration would attempt INSERTs that conflict with sequences already
-- past their old values, so we don't try to resurrect them.
SELECT 1;
