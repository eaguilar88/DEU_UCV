DELETE FROM deu.user_roles ur USING deu.users u
WHERE
  ur.user_id = u.id
  AND u.ci IN (115, 116);

DELETE FROM deu.users
WHERE
  ci IN (115, 116);
