ALTER TABLE deu.courses ADD COLUMN origin_faculty deu.faculty_enum;
UPDATE deu.courses SET origin_faculty = faculty;
