DELETE FROM deu.files WHERE owner_type = 'group_renewal';

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

-- Without renewal_id these would look like creation requests.
DELETE FROM deu.group_auth_requests WHERE renewal_id IS NOT NULL;
ALTER TABLE deu.group_auth_requests DROP COLUMN renewal_id;

DROP TABLE IF EXISTS deu.group_renewals;
