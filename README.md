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
make start-db
```

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