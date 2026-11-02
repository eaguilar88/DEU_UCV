# Guía para probar diplomados con el backend Go

Esta guía levanta en tu computadora el frontend de diplomados (Next.js) conectado al backend Go, con su base de datos y servicios auxiliares, y propone un recorrido para probar el flujo completo de un curso.

## 1. Requisitos

Instalado en tu máquina:

- **Git**, con acceso por SSH a GitHub a los dos repositorios:
  - `git@github.com:eaguilar88/DEU_UCV.git` (backend Go)
  - `git@github.com:RoaRobinson97/ecp_ucv.git` (frontend diplomados, submódulo de `DEU_UCV`)
- **Docker** con **Docker Compose v2** (el comando es `docker compose`, sin guion).
- **make**.
- **Node.js 20** (el mismo de la imagen Docker del frontend) y **npm**.
- El archivo **`.env`** que te compartieron. Tiene credenciales reales (correo, Backblaze B2): no lo subas a git ni lo compartas.

No necesitas Go instalado: el backend corre dentro de Docker.

## 2. Obtener el código

```bash
git clone git@github.com:eaguilar88/DEU_UCV.git
cd DEU_UCV
git checkout feature/nuevos-correos-grupos

# El frontend es un submódulo; se descarga y se cambia a su rama de trabajo
git submodule update --init diplomados
cd diplomados
git checkout feature/backend-go
git pull
cd ..
```

Copia el `.env` que te compartieron en la raíz de `DEU_UCV/` (junto al `Makefile`).

## 3. Levantar el backend Go

Todos los comandos se ejecutan desde la raíz de `DEU_UCV/`.

```bash
make start-backend   # construye y levanta backend, notifier, gotenberg y la base de datos
make migrate-up      # crea las tablas y los usuarios de prueba
```

Corre `make migrate-up` la primera vez y cada vez que actualices el código: las migraciones nuevas no se aplican solas. Si el backend arrancó antes de que existieran las tablas, reinícialo con `docker compose restart backend`.

El backend queda en http://localhost:8081. Para comprobarlo:

```bash
docker compose ps                  # los servicios deben estar "running"
docker compose logs -f backend     # logs del backend (Ctrl+C para salir)
```

Servicios que se levantan (`docker-compose.yml`):

| Servicio    | Para qué sirve                                                                  |
| ----------- | ------------------------------------------------------------------------------- |
| `backend`   | API Go en el puerto 8081.                                                       |
| `db`        | PostgreSQL 16 (contenedor `pg16`).                                              |
| `gotenberg` | Convierte HTML a PDF para los certificados. Solo lo usa el backend.             |
| `notifier`  | Envía los correos de recordatorio al arrancar y luego una vez al día.           |

### Comandos útiles

| Comando                                   | Qué hace                                                                             |
| ----------------------------------------- | ------------------------------------------------------------------------------------ |
| `make start-backend`                      | Construye las imágenes y levanta todo. Úsalo también después de actualizar el código. |
| `make stop-backend`                       | Detiene todo. Los datos de la base se conservan.                                     |
| `make migrate-up`                         | Aplica las migraciones pendientes.                                                   |
| `make migrate-version`                    | Muestra la última migración aplicada.                                                |
| `docker compose logs -f backend`          | Sigue los logs del backend. Cambia `backend` por `notifier`, `db` o `gotenberg`.     |
| `docker compose restart backend`          | Reinicia solo el backend.                                                            |
| `docker compose stop notifier`            | Detiene los recordatorios automáticos si no quieres recibirlos.                      |
| `docker compose down -v`                  | Detiene todo **y borra la base de datos**. Después: `make start-backend` y `make migrate-up`. |

No uses `make start-db`, `make stop-db` ni `make refresh`: usan otro archivo (`docker-compose.dev.yml`) pensado para correr Go fuera de Docker, y chocan con el contenedor `pg16` de esta guía. `make stop-db` además borra la base.

## 4. Levantar el frontend de diplomados

En otra terminal, dentro de `DEU_UCV/diplomados/`:

```bash
npm install                          # solo la primera vez o si cambia package.json
cp .env.local.example .env.local     # solo la primera vez: INTERNAL_API_URL=http://localhost:8081
npm run dev
```

Abre http://localhost:3000.

El navegador habla con `/api/*` del propio frontend, y Next.js reenvía esas peticiones al backend Go. Por eso no hace falta configurar nada más. Si la página muestra errores de conexión, revisa que el backend esté arriba (`docker compose ps`).

## 5. Usuarios de prueba

Las migraciones crean estos usuarios. Todos tienen la contraseña **`nolodire`** y se inicia sesión con el correo.

| Rol                          | Correo                                     | Para qué usarlo                                          |
| ---------------------------- | ------------------------------------------ | -------------------------------------------------------- |
| Administrador DEU            | `deu_admin1@example.com`                   | Ver todas las solicitudes, de todas las facultades.      |
| Coordinador de Ciencias      | `coordinador_ciencias@extension.ucv.ve`    | Revisar proveedores y cursos de la Facultad de Ciencias. |
| Coordinador de Odontología   | `coordinador_odontologia@extension.ucv.ve` | Recibir cursos remitidos desde Ciencias.                 |

Hay un coordinador por facultad, con el mismo formato de correo: `coordinador_ingenieria@extension.ucv.ve`, `coordinador_medicina@extension.ucv.ve`, etc.

El **proveedor** lo creas tú (ver §6).

## 6. Recorrido sugerido

1. **Regístrate** en http://localhost:3000/registro con un correo tuyo.
2. **Solicita ser proveedor** desde "Solicitar organización" (`/solicitar-organizacion`). Elige la facultad **Ciencias**.
3. **Aprueba al proveedor:** cierra sesión, entra como `coordinador_ciencias@extension.ucv.ve` y ve a "Panel de Administración" (`/admin`).
4. **Formula un curso:** vuelve a entrar con tu usuario. Debes **cerrar sesión y volver a entrar** después de la aprobación: el código de proveedor viaja en el token de sesión, y hasta que no inicies sesión de nuevo no aparece la opción "Formular" (`/formulacion-de-curso`).
5. **Revisa el curso** como coordinador de Ciencias: apruébalo o recházalo. También puedes **remitirlo** a Odontología y revisarlo como `coordinador_odontologia@extension.ucv.ve`.
6. **Ampara el curso:** como coordinador de Ciencias, desde el perfil del proveedor, carga el contrato (carta de intención y carta de compromiso). Sin este paso el curso no puede abrir cohorte. Después de ampararlo, el curso ya aparece en el catálogo público (`/cursos`).
7. **Abre una cohorte** como proveedor: en "Mis cursos" (`/mis-cursos`) entra al curso y usa la gestión de cohorte de su página (`/curso/<id>`).
8. **Solicita el cierre de la cohorte** como proveedor, desde la misma página del curso. Pide tres archivos: participantes, vouchers y encuesta.
9. **Aprueba o rechaza el cierre** como coordinador de Ciencias. Las solicitudes de cierre van siempre a la facultad de origen del curso, aunque se haya remitido.

## 7. Consejos

- **Usa un correo al que tengas acceso** para registrarte: así ves las notificaciones que recibe el proveedor (solicitud aprobada, curso aprobado, cierre aprobado, etc.). Los correos se envían de verdad.
- **Los coordinadores de prueba tienen correos ficticios**, así que no verás lo que reciben. Si quieres verlo, cambia el correo de uno por otro tuyo (desde ese momento inicias sesión con el correo nuevo):
  ```bash
  docker exec -it pg16 psql -U deu_admin -d deu \
    -c "UPDATE deu.users SET email = 'tu.otro.correo@ejemplo.com' WHERE email = 'coordinador_ciencias@extension.ucv.ve';"
  ```
  Los valores de `-U` y `-d` son `POSTGRES_USER` y `POSTGRES_DB` del `.env`.
- **Puedes subir el mismo archivo varias veces.** Un PDF cualquiera sirve para todos los adjuntos (cartas, CV, participantes, vouchers, encuesta), y una imagen JPG o PNG para la portada del curso.
- **Si una aprobación o un rechazo falla con un error del servidor**, revisa `docker compose logs backend`. Si el correo de notificación no se pudo enviar, la acción se revierte a propósito, para que nadie quede sin avisar.
- **Para empezar de cero:** `docker compose down -v`, luego `make start-backend` y `make migrate-up`. Se borran todos los datos, incluido tu usuario.
- **Después de actualizar el código** (`git pull` en `DEU_UCV` y en `diplomados`), corre `make start-backend`, `make migrate-up` y, si cambió `package.json`, `npm install`.
