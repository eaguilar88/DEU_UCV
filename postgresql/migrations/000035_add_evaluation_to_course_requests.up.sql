-- Evaluation recorded when a course request is approved. The evaluation document is a file
-- (deu.files, owner_type 'course_request', purpose 'archivo_evaluacion').
ALTER TABLE deu.course_auth_requests ADD COLUMN score NUMERIC(5, 2);
ALTER TABLE deu.course_auth_requests ADD COLUMN classification TEXT;

-- Recreate the owner_type check to allow 'course_request'. The original list (000017) lacked a
-- comma after 'group_member', which concatenated it with 'course_cycle_close_request' and
-- rejected both values.
ALTER TABLE deu.files DROP CONSTRAINT IF EXISTS files_owner_type_check;
ALTER TABLE deu.files ADD CONSTRAINT files_owner_type_check CHECK (
  owner_type IN (
    'course',
    'group',
    'group_activity',
    'course_cycle',
    'provider',
    'user',
    'group_member',
    'course_cycle_close_request',
    'provider_contract',
    'course_request'
  )
);
