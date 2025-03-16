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