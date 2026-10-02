# Fuera de alcance

Este documento lista funcionalidades que se evaluaron y quedaron **fuera del alcance** de este proyecto. No son brechas pendientes: son decisiones de producto tomadas a propósito. Si alguna se retoma, debe planificarse como un proyecto aparte.

## 1. Manejo de participantes de cursos

### Decisión

La plataforma **no** registra a los participantes de los cursos. Los participantes no tienen acceso al sitio y toda la comunicación con ellos ocurre por canales externos (correo, mensajería, etc.), gestionada por el facilitador del curso.

El backend de Node (`diplomados/`) tampoco modelaba participantes, así que esto no deja ninguna brecha frente al mock (ver §5.1 del [análisis de brechas](gap_analysis_diplomados_node_mock.md)). Por la misma razón se eliminó el campo huérfano `Participants` de `entities.CoursePeriod`, que no tenía ningún endpoint ni persistencia asociados.

### Qué queda fuera

Queda fuera **registrar explícitamente en la plataforma a los participantes de un curso dado**, con el fin de tener sus datos disponibles en el sistema y generar los certificados a partir de ellos.

En su lugar, la estrategia adoptada es que **el facilitador adjunte un archivo** con los participantes que obtienen la certificación (hoy, `archivo_participantes` en la solicitud de cierre de cohorte), y los certificados se generan a partir de ese archivo.

### Comparación de estrategias

| Aspecto                         | Registro de participantes en la plataforma                                                           | Archivo adjunto por el facilitador (estrategia adoptada)                                               |
| ------------------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Datos disponibles para la DEU   | Todos los inscritos en cada cohorte, de forma estructurada                                           | Solo los participantes que el facilitador incluye (normalmente, los que se certifican)                 |
| Estadísticas                    | Permite a la DEU sacar estadísticas (demografía, deserción, aprobación, etc.)*                       | Limitadas a lo que contenga el archivo                                                                 |
| Privacidad                      | Centraliza datos personales de todos los estudiantes; exige políticas de resguardo y consentimiento  | El facilitador decide qué datos comparte; puede enviar solo los de quienes obtienen la certificación   |
| Agilidad para el facilitador    | Requiere cargar/gestionar inscripciones en la plataforma durante todo el curso                       | Un único paso al cierre de la cohorte                                                                  |
| Generación de certificados      | Directa desde los datos registrados                                                                  | A partir del archivo (requiere un formato de archivo acordado)                                         |
| Acceso de participantes al sitio| Normalmente implica cuentas y autenticación para participantes                                       | No requiere cuentas para participantes                                                                 |
| Costo de implementación         | Alto: inscripción, gestión de usuarios participantes, comunicación, protección de datos              | Bajo: reutiliza la carga de archivos existente                                                         |

\* Si la ley lo permite. Determinar el marco legal aplicable al tratamiento de datos personales de los participantes también queda fuera del alcance de este proyecto.
