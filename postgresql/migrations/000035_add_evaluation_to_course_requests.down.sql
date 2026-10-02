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
    'provider_contract'
  )
);

ALTER TABLE deu.course_auth_requests DROP COLUMN classification;
ALTER TABLE deu.course_auth_requests DROP COLUMN score;
