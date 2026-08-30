-- Минцифры catalogue contains product names longer than the original 256-char
-- VARCHAR (e.g. id=23196). Switch to TEXT so the importer doesn't have to
-- truncate.

ALTER TABLE software_catalog
    ALTER COLUMN name   TYPE TEXT,
    ALTER COLUMN vendor TYPE TEXT;
