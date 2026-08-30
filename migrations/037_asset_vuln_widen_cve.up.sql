-- БДУ-запись часто перечисляет десятки связанных CVE (snapshot хранит их
-- через пробел: "CVE-2014-0224 CVE-2014-0160 …"). VARCHAR(64) не вмещает
-- такой список — расширяем поле до TEXT.

ALTER TABLE asset_vulnerabilities
    ALTER COLUMN cve TYPE TEXT;
