# PROYECTO DEU UCV
Sistema para la Dirección de Extensión Universitaria de la UCV

## Alcance
* Educación Continua y Permanente 
    * Facilitación del Proceso de solicitud y aprobación de avales de cursos de extensión.
    * Publicación de la información de los cursos.
    * Control de ejecución de cohortes de los cursos.
* Gestion Social Universitaria
    * Registro de grupos de extensión
    * Actividades de los grupos de extensión
## Responsables
- Ellery Aguilar
- Robinson Roa
- Sonia Nahit
- Kristian Ferreira
- Ronald More

## Enlaces a otros proyectos
- [Landing del proyecto](https://github.com/soniamira/deuweb)
- [Nueva landing (Rails)](https://github.com/danielitomoros03/deu-app)
- [Sistema de diplomados](https://github.com/RoaRobinson97/ecp_ucv)

## Tecnologías principales del proyecto
- Golang 1.22+
- Docker
- PostgreSQL 16 
- Blackblaze B2
- Traefik
### Librerías principales para el servidor
- [Echo Web Framework](https://echo.labstack.com/docs): Framework principal para organizar todo el servidor HTTP.
- [Squirrel](https://github.com/Masterminds/squirrel): Librería usada para escribir SQL de manera sencilla.
- [Zap](https://github.com/uber-go/zap): Librería para crear e imprimir logs estructurados.
- [Amazon Golang SDK](https://github.com/aws/aws-sdk-go-v2): Librería para acceder a servicios de almacenamiento en la nube. Necesaria para [conectarse a Blackblaze B2](https://www.backblaze.com/docs/cloud-storage-use-the-aws-sdk-for-go-with-backblaze-b2).
- [golang-migrate](https://github.com/golang-migrate/migrate): Herramienta para gestionar las migraciones del esquema de PostgreSQL, ejecutada como servicio Docker (ver [postgresql/README.md](postgresql/README.md)).

## Herramientas para la documentación
- [DBML(Database Markup Language)](https://docs.dbdiagram.io/): Lenguaje utilizado para definir programáticamente el modelo de la base de datos.
- [Lucidchart](https://www.lucidchart.com/pages/es): Para el resto de diagramas.
## Iniciar desarrollo

Para renombrar el archivo `.env.example` a `.env`, puede usar uno de los siguientes comandos:

Usando `cp`:
```
cp .env.example .env
```

Usando `mv`:
```
mv .env.example .env
```

### Desarrollo Backend
Para iniciar el desarrollo del backend, ejecute el siguiente comando:
```
make start-backend
```

Luego, aplique las migraciones de la base de datos (crea el esquema `deu` y los datos semilla básicos):
```
make migrate-up
```

Para más detalles sobre el manejo de migraciones (crear nuevas, revertir, cargar datos de prueba opcionales), vea [postgresql/README.md](postgresql/README.md).

### Desarrollo Frontend
Para iniciar el desarrollo del frontend, ejecute el siguiente comando:
```
make start-backend
```

### Entorno de Producción
Para iniciar el entorno de producción, ejecute el siguiente comando:
```
make start-prod
```

En un primer despliegue, aplique también las migraciones de la base de datos:
```
make migrate-up-prod
```

Vea [postgresql/README.md](postgresql/README.md) para más detalles.

Si desea desplegar solo la nueva landing Rails y sus dependencias directas (sin levantar `diplomados` y `grupos`), use:

```
make start-prod-landing
```

## Despliegue de la nueva landing Rails (deu-app)

`docker-compose.prod.yml` está configurado para construir el servicio `landing` desde la carpeta local `landing/`.

La diferencia es solo el origen del submódulo: ahora `landing/` apunta al repositorio `deu-app`.

### Pasos recomendados de despliegue

1. Crear `.env` a partir de `.env.example`.
2. Definir `SECRET_KEY_BASE` con un valor seguro y aleatorio.
3. Inicializar/actualizar solo el submódulo `landing`:

```
git submodule sync -- landing
git submodule update --init --recursive landing
```

4. Levantar producción:

```
make start-prod-landing
```

5. Ejecutar seed de Rails (usuarios base y contenido inicial):

```
make seed-landing
```

### Variables relevantes para la landing Rails

- `SECRET_KEY_BASE`: obligatorio para producción.
- `RAILS_FORCE_SSL`: activar solo si tienes HTTPS terminado antes de la app/proxy.
- `DEU_SEED_PASSWORD`: contraseña para usuarios semilla.
- `DEU_SUPER_ADMIN_EMAILS`: lista de correos de superadmin separados por coma.

### Notas operativas

- El servicio `landing` usa PostgreSQL del servicio `db` vía `DATABASE_URL`.
- El almacenamiento local de Active Storage persiste en el volumen `landing_storage`.
- El seed es idempotente y se puede ejecutar más de una vez.