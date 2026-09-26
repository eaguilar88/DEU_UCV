-- The /files proxy only serves files flagged as public. Uploads never persisted the flag
-- before this migration, so backfill it for the file purposes meant to be public.
UPDATE deu.files
SET public = true
WHERE (owner_type = 'course' AND purpose = 'portada')
   OR (owner_type = 'group_activity' AND purpose = 'cubierta')
   OR (owner_type IN ('group', 'provider') AND purpose = 'logo');

CREATE INDEX IF NOT EXISTS idx_files_file_key ON deu.files (file_key);
