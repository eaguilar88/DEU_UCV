DELETE FROM deu.users
WHERE
  email IN (
    'eaguilar@email.com',
    'deu_admin1@example.com',
    'deu_admin2@example.com',
    'coordinador_agronomia@extension.ucv.ve',
    'coordinador_arquitectura@extension.ucv.ve',
    'coordinador_ciencias@extension.ucv.ve',
    'coordinador_fases@extension.ucv.ve',
    'coordinador_ciencias_politicas@extension.ucv.ve',
    'coordinador_veterinaria@extension.ucv.ve',
    'coordinador_farmacia@extension.ucv.ve',
    'coordinador_humanidades@extension.ucv.ve',
    'coordinador_ingenieria@extension.ucv.ve',
    'coordinador_medicina@extension.ucv.ve',
    'coordinador_odontologia@extension.ucv.ve'
  );
