DROP VIEW IF EXISTS ptszi_ubi_source_summary;
DELETE FROM ptszi_threat_ubi_links
WHERE comment = 'Автоматическая первичная классификация УБИ по тексту БДУ ФСТЭК; используется как справочная связь с ST.';
DROP TABLE IF EXISTS ptszi_ubi_source_mappings;
DROP TABLE IF EXISTS ptszi_ubi_threats;
