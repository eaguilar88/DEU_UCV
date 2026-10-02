-- Course proposal fields captured by the diplomados form: the modules with their contents and
-- competencies, and the bibliography. The facilitator CV is a file (deu.files, purpose
-- 'cv_facilitador').
ALTER TABLE deu.courses ADD COLUMN competencies TEXT;
ALTER TABLE deu.courses ADD COLUMN bibliography TEXT;
