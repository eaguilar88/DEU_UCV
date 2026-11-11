DELETE FROM deu.user_roles ur USING deu.users u
WHERE
  ur.user_id = u.id
  AND u.ci = 117;

DELETE FROM deu.users
WHERE
  ci = 117;
