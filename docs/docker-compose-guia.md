# Guía para Agregar Servicios a Docker Compose (Producción)

Esta guía te ayudará a agregar nuevos servicios al archivo `docker-compose.prod.yml` del proyecto DEU_UCV.

---

## 1. Cómo Agregar un Servicio

Puedes agregar diferentes tipos de servicios según las necesidades del proyecto:

- **Bases de datos** (PostgreSQL, Redis, MongoDB, etc.)
- **Aplicaciones frontend** (React, Vue, Angular, etc.)
- **Aplicaciones backend** (Go, Node.js, Python, etc.)

### Estructura Básica de un Servicio

```yaml
services:
  nombre-del-servicio:
    image: nombre-imagen:version
    container_name: nombre-contenedor
    restart: unless-stopped
    env_file:
      - .env
    networks:
      - web
    depends_on:
      - otro-servicio
```

### Aspectos Clave

#### 1.1 Network (Red)

**Todos los servicios deben estar en la misma red** para poder comunicarse entre sí.

```yaml
networks:
  - web
```

**¿Por qué es importante?**
- Permite que los contenedores se comuniquen usando sus nombres de servicio
- Ejemplo: El backend puede conectarse a PostgreSQL usando `db:5432` en lugar de una IP

**Al final del archivo**, asegúrate de que la red esté definida:

```yaml
networks:
  web:
    driver: bridge
```

#### 1.2 depends_on (Dependencias)

Define el **orden de inicio** de los servicios. Un servicio esperará a que sus dependencias estén "listas" antes de iniciar.

```yaml
depends_on:
  - db
  - backend
```

**Ejemplo práctico:**
```yaml
services:
  landing:
    # ...
    depends_on:
      - backend  # El frontend necesita que el backend esté corriendo primero
      - db       # Y también la base de datos
```

**⚠️ Importante:** `depends_on` solo espera que el contenedor esté corriendo, **no** que el servicio interno esté listo. Para bases de datos, considera usar health checks.

**Ejemplo con health check:**
```yaml
services:
  backend:
    depends_on:
      db:
        condition: service_healthy  # Espera a que db esté "healthy"

  db:
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $POSTGRES_USER"]
      interval: 5s
      timeout: 5s
      retries: 5
```

#### 1.3 env_file (Variables de Entorno)

Carga las variables de entorno desde el archivo `.env`:

```yaml
env_file:
  - .env
```

**¿Qué contiene el `.env`?**
- Credenciales de base de datos
- API keys
- URLs de servicios externos
- Configuración específica del ambiente

**Ejemplo de `.env`:**
```bash
POSTGRES_USER=deu_user
POSTGRES_PASSWORD=secret123
POSTGRES_DB=deu_db
API_PORT=8080
```

### Ejemplos de Servicios Comunes

#### Ejemplo 1: Agregar Redis (Base de Datos)

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: deu-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    networks:
      - web
    volumes:
      - redis-data:/data

volumes:
  redis-data:
```

#### Ejemplo 2: Agregar Frontend (React/Vue)

```yaml
services:
  mi-frontend:
    build:
      context: ./mi-frontend
      dockerfile: Dockerfile
    container_name: deu-mi-frontend
    restart: unless-stopped
    env_file:
      - .env
    networks:
      - web
    depends_on:
      - backend
      - db
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.mi-frontend.rule=Host(`app.extension.ucv.ve`)"
      - "traefik.http.routers.mi-frontend.entrypoints=web"
      - "traefik.http.services.mi-frontend.loadbalancer.server.port=80"
```

#### Ejemplo 3: Agregar Otro Backend (Node.js/Python)

```yaml
services:
  api-notifications:
    build:
      context: ./notifications-service
      dockerfile: Dockerfile
    container_name: deu-notifications
    restart: unless-stopped
    env_file:
      - .env
    networks:
      - web
    depends_on:
      db:
        condition: service_healthy
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.notifications.rule=Host(`api.extension.ucv.ve`) && PathPrefix(`/notifications`)"
      - "traefik.http.routers.notifications.entrypoints=web"
      - "traefik.http.services.notifications.loadbalancer.server.port=3000"
```

---

## 2. Traefik: Reverse Proxy y Load Balancer

### ¿Qué es Traefik?

**Traefik** es un reverse proxy moderno que actúa como la "puerta de entrada" a tu aplicación. Recibe todas las peticiones HTTP/HTTPS y las redirige al servicio correcto.

**Funciones principales:**
- ✅ **Routing**: Dirige las peticiones al servicio correcto según el dominio o path
- ✅ **SSL/TLS**: Maneja certificados HTTPS automáticamente (Let's Encrypt)
- ✅ **Load Balancing**: Distribuye la carga entre múltiples instancias
- ✅ **Service Discovery**: Detecta automáticamente nuevos servicios

### Diagrama de Arquitectura

<!-- PLACEHOLDER: Insertar diagrama de Traefik aquí -->
<!-- El diagrama debe mostrar:
     - Cliente (navegador)
     - Traefik (reverse proxy)
     - Servicios (backend, landing, diplomados, grupos, db)
     - Red Docker (web)
-->

```
[Diagrama pendiente - Mostrar flujo: Cliente → Traefik → Servicios]
```

### Cómo Configurar Traefik para tu Servicio

Para que Traefik redirija peticiones a tu servicio, debes agregar **labels** (etiquetas) en la configuración del servicio.

#### Labels Esenciales

```yaml
labels:
  # 1. Habilitar Traefik para este servicio
  - "traefik.enable=true"
  
  # 2. Definir la regla de routing (cuándo usar este servicio)
  - "traefik.http.routers.NOMBRE-ROUTER.rule=Host(`dominio.com`)"
  
  # 3. Especificar el puerto interno del contenedor
  - "traefik.http.services.NOMBRE-SERVICIO.loadbalancer.server.port=PUERTO"
```

#### Ejemplo Real: Backend API

```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    env_file:
      - .env
    networks:
      - web
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`api.extension.ucv.ve`)"
      - "traefik.http.routers.backend.entrypoints=web"
      - "traefik.http.services.backend.loadbalancer.server.port=80"
```

**Explicación:**
- `traefik.enable=true` → Activa Traefik para este servicio
- `Host(\`api.extension.ucv.ve\`)` → Peticiones a `api.extension.ucv.ve` van a este servicio
- `entrypoints=web` → Usa HTTP (puerto 80)
- `server.port=80` → El contenedor escucha en el puerto 80 internamente
- `depends_on` con `condition: service_healthy` → Espera a que la base de datos esté lista

#### Routing por Path (Ruta)

Si quieres que diferentes rutas vayan a diferentes servicios:

```yaml
# Servicio 1: API principal
labels:
  - "traefik.http.routers.api.rule=Host(`api.extension.ucv.ve`) && PathPrefix(`/api`)"
  - "traefik.http.routers.api.entrypoints=web"
  - "traefik.http.services.api.loadbalancer.server.port=80"

# Servicio 2: Servicio de notificaciones
labels:
  - "traefik.http.routers.notifications.rule=Host(`api.extension.ucv.ve`) && PathPrefix(`/notifications`)"
  - "traefik.http.routers.notifications.entrypoints=web"
  - "traefik.http.services.notifications.loadbalancer.server.port=3000"
```

**Resultado:**
- `http://api.extension.ucv.ve/api/...` → Va al servicio principal (puerto 80)
- `http://api.extension.ucv.ve/notifications/...` → Va al servicio de notificaciones (puerto 3000)

#### Middleware (Opcional)

Puedes agregar middleware para funcionalidades adicionales:

```yaml
labels:
  # Redirección HTTP → HTTPS
  - "traefik.http.middlewares.redirect-to-https.redirectscheme.scheme=https"
  - "traefik.http.routers.backend-http.middlewares=redirect-to-https"

  # CORS (Cross-Origin Resource Sharing)
  - "traefik.http.middlewares.cors.headers.accesscontrolallowmethods=GET,POST,PUT,DELETE,OPTIONS"
  - "traefik.http.middlewares.cors.headers.accesscontrolalloworigin=*"
  - "traefik.http.routers.backend.middlewares=cors"

  # Rate Limiting (límite de peticiones)
  - "traefik.http.middlewares.ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.ratelimit.ratelimit.burst=50"
```

---

## 3. Troubleshooting y Testing

### 3.1 Verificar la Configuración

Antes de levantar los servicios, valida la sintaxis del archivo:

```bash
# Validar sintaxis de docker-compose
docker-compose -f docker-compose.prod.yml config
```

Si hay errores de sintaxis, este comando te los mostrará.

### 3.2 Levantar los Servicios

```bash
# Levantar todos los servicios
docker-compose -f docker-compose.prod.yml up -d

# Levantar un servicio específico
docker-compose -f docker-compose.prod.yml up -d nombre-servicio
```

### 3.3 Verificar el Estado de los Servicios

```bash
# Ver todos los contenedores corriendo
docker-compose -f docker-compose.prod.yml ps

# Ver logs de un servicio específico
docker-compose -f docker-compose.prod.yml logs -f nombre-servicio

# Ver logs de Traefik (para debugging de routing)
docker-compose -f docker-compose.prod.yml logs -f traefik
```

### 3.4 Problemas Comunes y Soluciones

#### Problema 1: El servicio no inicia

**Síntomas:**
```bash
docker-compose ps
# Estado: Restarting o Exit
```

**Solución:**
```bash
# Ver los logs del servicio
docker-compose -f docker-compose.prod.yml logs nombre-servicio

# Verificar que las dependencias estén corriendo
docker-compose -f docker-compose.prod.yml ps
```

**Causas comunes:**
- Variables de entorno faltantes en `.env`
- Dependencias (`depends_on`) no están corriendo
- Puerto ya en uso
- Error en el código de la aplicación

#### Problema 2: Traefik no redirige las peticiones

**Síntomas:**
- Error 404 o "Service Unavailable"
- La petición no llega al servicio

**Solución:**
```bash
# 1. Verificar que Traefik detectó el servicio
docker-compose -f docker-compose.prod.yml logs traefik | grep "nombre-servicio"

# 2. Verificar los labels del servicio
docker inspect nombre-contenedor | grep -A 20 Labels

# 3. Acceder al dashboard de Traefik (si está habilitado)
# http://localhost:8080 o https://traefik.deu.ucv.ve
```

**Checklist:**
- ✅ `traefik.enable=true` está presente
- ✅ El servicio está en la red `web`
- ✅ El puerto especificado en `loadbalancer.server.port` es correcto
- ✅ La regla de routing (`Host` o `PathPrefix`) es correcta
- ✅ El `entrypoints` está configurado (normalmente `web` para HTTP)
- ✅ El DNS apunta al servidor correcto

#### Problema 3: Error de red entre servicios

**Síntomas:**
- El backend no puede conectarse a la base de datos
- Error: "connection refused" o "host not found"

**Solución:**
```bash
# 1. Verificar que ambos servicios están en la misma red
docker network inspect web

# 2. Probar conectividad desde un contenedor
docker exec -it nombre-contenedor ping otro-servicio

# 3. Verificar variables de entorno
docker exec -it nombre-contenedor env | grep POSTGRES
```

**Checklist:**
- ✅ Ambos servicios tienen `networks: - web`
- ✅ Usas el nombre del servicio (ej: `db`, no `localhost`) para conectar
- ✅ El puerto es el interno del contenedor, no el expuesto
- ✅ Ejemplo: Para conectar a PostgreSQL usa `db:5432`, no `localhost:5432`

#### Problema 4: Certificado SSL no se genera

**Síntomas:**
- Error "certificate not found"
- Navegador muestra "conexión no segura"

**Solución:**
```bash
# Ver logs de Traefik para errores de Let's Encrypt
docker-compose -f docker-compose.prod.yml logs traefik | grep -i "acme\|certificate"
```

**Checklist:**
- ✅ El dominio apunta a la IP del servidor
- ✅ Los puertos 80 y 443 están abiertos en el firewall
- ✅ `certresolver=letsencrypt` está configurado
- ✅ El email en la configuración de Traefik es válido

### 3.5 Testing Manual

#### Test 1: Verificar que el servicio responde

```bash
# Desde el servidor (puerto expuesto)
curl http://localhost:8081/health

# Desde fuera (con Traefik)
curl http://api.extension.ucv.ve/health
```

#### Test 2: Verificar headers de respuesta

```bash
# Ver headers completos
curl -I http://api.extension.ucv.ve/api/courses

# Verificar CORS
curl -H "Origin: http://formacion.extension.ucv.ve" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS \
     http://api.extension.ucv.ve/api/courses
```

#### Test 3: Verificar conectividad entre servicios

```bash
# Entrar al contenedor del backend
docker exec -it backend sh

# Probar conexión a PostgreSQL (usa el nombre del servicio: db)
nc -zv db 5432

# O con ping
ping db

# Verificar que la variable de entorno apunta al servicio correcto
echo $POSTGRES_HOST  # Debe ser "db", no "localhost"
```

### 3.6 Comandos Útiles

```bash
# Reiniciar un servicio específico
docker-compose -f docker-compose.prod.yml restart nombre-servicio

# Reconstruir y reiniciar un servicio
docker-compose -f docker-compose.prod.yml up -d --build nombre-servicio

# Detener todos los servicios
docker-compose -f docker-compose.prod.yml down

# Detener y eliminar volúmenes (⚠️ CUIDADO: borra datos)
docker-compose -f docker-compose.prod.yml down -v

# Ver uso de recursos
docker stats

# Limpiar recursos no usados
docker system prune -a
```

---

## 4. Cómo Desplegar tu Proyecto en el VPS del Equipo

Esta sección está pensada para quienes quieren probar su propio proyecto usando el servidor compartido del equipo, sin necesidad de tener un dominio propio.

La idea es simple: cada persona expone su aplicación en un **puerto diferente**, y se accede a ella usando la dirección IP del servidor seguida de ese puerto. Por ejemplo: `http://123.456.789.0:8082`.

---

### Paso 1: Conectarte al servidor por SSH

SSH es una forma de conectarte a un servidor remoto desde tu computadora, como si abrieras una terminal directamente en ese servidor.

Abre una terminal en tu computadora y ejecuta:

```bash
ssh usuario@IP_DEL_SERVIDOR
```

Reemplaza `usuario` con el nombre de usuario que te dieron, e `IP_DEL_SERVIDOR` con la dirección IP del VPS. Te va a pedir la contraseña — escríbela y presiona Enter (no vas a ver los caracteres mientras escribes, eso es normal).

**Ejemplo:**
```bash
ssh maria@123.456.789.0
```

Si es la primera vez que te conectas, va a aparecer un mensaje preguntando si confías en el servidor. Escribe `yes` y presiona Enter.

Una vez conectada, vas a ver algo como esto en tu terminal:
```
maria@vps-servidor:~$
```

Eso significa que ya estás dentro del servidor. Todo lo que escribas desde ahora se ejecuta ahí.

---

### Paso 2: Crear una carpeta para tu proyecto

Es importante que cada persona trabaje en su propia carpeta para no mezclar archivos con los demás.

```bash
mkdir ~/mi-proyecto
cd ~/mi-proyecto
```

Esto crea una carpeta llamada `mi-proyecto` en tu directorio personal y entra a ella. Puedes cambiar el nombre por el que quieras.

---

### Paso 3: Coordinar qué puerto vas a usar

Cada proyecto debe usar un puerto diferente. Si dos personas usan el mismo puerto, una de las dos no va a poder levantar su servicio.

Consulta con el equipo qué puertos ya están en uso. Como referencia, estos ya están ocupados por el proyecto DEU:
- `80` → Traefik (HTTP)
- `443` → Traefik (HTTPS)
- `8080` → Dashboard de Traefik
- `8081` → Backend de DEU

Para proyectos de prueba del equipo, usen el rango **8082 en adelante**. Pongan su nombre y puerto en un documento compartido para evitar conflictos.

---

### Paso 4: Crear tu archivo docker-compose.yml

Dentro de tu carpeta, crea el archivo de configuración de Docker:

```bash
nano docker-compose.yml
```

Esto abre un editor de texto en la terminal. Escribe (o pega) la configuración de tu proyecto. Aquí hay un ejemplo con una app de Node.js:

```yaml
services:
  mi-app:
    image: node:20-alpine        # La imagen de Docker que necesitas
    container_name: mi-app       # Un nombre único para tu contenedor
    restart: unless-stopped
    ports:
      - "8082:3000"              # Puerto del servidor : puerto interno de tu app
    working_dir: /app
    volumes:
      - ./:/app                  # Monta tu código en el contenedor
    command: node index.js
```

**Lo más importante aquí es la línea `ports`:**
- El número de la **izquierda** (`8082`) es el puerto del servidor — el que coordinaste con el equipo.
- El número de la **derecha** (`3000`) es el puerto que usa tu aplicación internamente — depende de cómo esté configurada tu app.

Para guardar en nano: presiona `Ctrl + X`, luego `Y`, luego `Enter`.

---

### Paso 5: Subir tu código al servidor

Tienes dos opciones para tener tu código en el servidor:

**Opción A — Clonar desde Git (recomendado):**
```bash
git clone https://github.com/tu-usuario/tu-repositorio.git .
```

El `.` al final clona el repositorio en la carpeta actual.

**Opción B — Copiar archivos desde tu computadora:**

Abre **otra terminal** en tu computadora (no la que tiene el SSH) y ejecuta:
```bash
scp -r ./mi-proyecto/* usuario@IP_DEL_SERVIDOR:~/mi-proyecto/
```

Esto copia todos los archivos de tu carpeta local al servidor.

---

### Paso 6: Levantar tu proyecto

Ya con el código y el `docker-compose.yml` listos, levanta los contenedores:

```bash
docker compose up -d
```

La flag `-d` hace que los contenedores corran en background (en segundo plano), así no bloquean tu terminal.

Para verificar que está corriendo:
```bash
docker compose ps
```

Deberías ver tu contenedor con el estado `running`.

---

### Paso 7: Acceder a tu aplicación

Abre el navegador y ve a:

```
http://IP_DEL_SERVIDOR:TU_PUERTO
```

Por ejemplo: `http://123.456.789.0:8082`

Si ves tu aplicación, ¡listo! Ya está funcionando.

---

### Comandos útiles del día a día

```bash
# Ver si tus contenedores están corriendo
docker compose ps

# Ver los logs de tu app (útil para ver errores)
docker compose logs -f

# Detener tu proyecto
docker compose down

# Reiniciar tu proyecto (por ejemplo, después de cambiar código)
docker compose down && docker compose up -d

# Ver los últimos 50 líneas de logs
docker compose logs --tail=50
```

---

### Problemas comunes

**"Port is already allocated" al levantar el contenedor**
Alguien más ya está usando ese puerto. Elige otro número en el `docker-compose.yml` y vuelve a intentarlo.

**No puedo acceder a `http://IP:PUERTO` desde el navegador**
Verifica que el contenedor esté corriendo con `docker compose ps`. Si el estado no dice `running`, revisa los logs con `docker compose logs` para ver el error.

**Me desconecté del SSH y el contenedor se detuvo**
Eso no debería pasar si usaste `-d` en el `docker compose up`. Verifica que hayas incluido `restart: unless-stopped` en tu `docker-compose.yml`.

---

## 5. Recursos Adicionales

### Documentación Oficial

- **Docker Compose**: https://docs.docker.com/compose/
- **Traefik**: https://doc.traefik.io/traefik/
- **Docker Networks**: https://docs.docker.com/network/

### Tutoriales Recomendados

- **Docker Compose para Producción**: https://docs.docker.com/compose/production/
- **Traefik con Let's Encrypt**: https://doc.traefik.io/traefik/https/acme/
- **Docker Networking Deep Dive**: https://docs.docker.com/network/bridge/

### Herramientas Útiles

- **Portainer**: Interfaz gráfica para gestionar Docker
  ```bash
  docker run -d -p 9000:9000 --name portainer \
    -v /var/run/docker.sock:/var/run/docker.sock \
    portainer/portainer-ce
  ```

- **Traefik Dashboard**: Visualiza rutas y servicios
  ```yaml
  # Agregar a traefik en docker-compose.prod.yml
  labels:
    - "traefik.http.routers.dashboard.rule=Host(`traefik.extension.ucv.ve`)"
    - "traefik.http.routers.dashboard.service=api@internal"
    - "traefik.http.routers.dashboard.entrypoints=web"
  ```

  **Nota**: El dashboard ya está habilitado en el puerto 8080 en la configuración actual.

- **ctop**: Monitor de contenedores en tiempo real
  ```bash
  docker run --rm -ti -v /var/run/docker.sock:/var/run/docker.sock quay.io/vektorlab/ctop:latest
  ```

### Comunidad y Soporte

- **Docker Community Forums**: https://forums.docker.com/
- **Traefik Community Forum**: https://community.traefik.io/
- **Stack Overflow**: Etiquetas `docker-compose`, `traefik`

---

## Checklist Final

Antes de hacer deploy a producción, verifica:

- [ ] Todos los servicios tienen `restart: unless-stopped`
- [ ] Las variables sensibles están en `.env` (no hardcodeadas)
- [ ] Todos los servicios están en la red `web`
- [ ] Los `depends_on` están correctamente configurados
- [ ] Los labels de Traefik son correctos (especialmente `entrypoints=web`)
- [ ] Los nombres de host en las reglas de Traefik usan `extension.ucv.ve`
- [ ] Los logs no muestran errores críticos
- [ ] Los health checks funcionan (especialmente para `db`)
- [ ] Los backups de base de datos están configurados
- [ ] El firewall permite tráfico en puertos 80 y 443
- [ ] Las secrets están configuradas correctamente (ej: `b2_key_id`, `b2_application_id`)
- [ ] Los volúmenes persistentes están definidos para datos importantes

---

---

## 6. Agregar Proyectos Independientes (Frontend + Backend + DB propios)

Esta sección cubre el caso donde se integra un proyecto completo nuevo al compose: con su propio frontend, su propio backend y opcionalmente su propia base de datos. Aplica para proyectos Node.js+Vue, Ruby+React, etc.

### Principios

- El **frontend** se expone vía Traefik con su propio DNS.
- El **backend** NO se expone vía Traefik — solo es accesible internamente por nombre de servicio Docker.
- El frontend se comunica con su backend usando la URL interna (configurada como variable de entorno en el build o en runtime).
- Si el proyecto necesita base de datos aislada, se agrega un contenedor `db-<proyecto>` con su propio volumen. Si puede compartir el Postgres existente, se usa una base de datos distinta dentro del mismo contenedor (`POSTGRES_DB`).
- Todos los servicios van en la red `web` para poder comunicarse entre sí.

### Template

```yaml
# === Proyecto: "proyecto-x" ===

  proyecto-x-frontend:
    build:
      context: "./proyecto-x/frontend"
      dockerfile: "Dockerfile"
    restart: unless-stopped
    depends_on:
      - proyecto-x-api
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.proyecto-x.rule=Host(`proyecto-x.extension.ucv.ve`)"
      - "traefik.http.routers.proyecto-x.entrypoints=web"
      - "traefik.http.services.proyecto-x.loadbalancer.server.port=80"
    networks:
      - web

  proyecto-x-api:
    build:
      context: "./proyecto-x/backend"
      dockerfile: "Dockerfile"
    restart: unless-stopped
    depends_on:
      db:                        # o db-proyecto-x si usa base de datos propia
        condition: service_healthy
    env_file:
      - ./proyecto-x/.env
    environment:
      DATABASE_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/proyecto_x
    labels:
      - "traefik.enable=false"   # No exponer públicamente
    networks:
      - web
```

### Cómo conecta el frontend con su backend

El frontend (Vue/React) hace sus llamadas API a una URL configurada en tiempo de build o como variable de entorno. Esa URL debe apuntar al nombre de servicio interno del backend y su puerto:

```
http://proyecto-x-api:3000
```

Esto funciona porque ambos contenedores están en la misma red Docker (`web`). Desde afuera del compose esa URL no es accesible — solo entre contenedores.

Si el frontend es una SPA (Single Page Application) que corre en el navegador del usuario, la comunicación no pasa por la red Docker. En ese caso el backend **sí necesita un dominio público** y debe exponerse vía Traefik:

```yaml
  proyecto-x-api:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.proyecto-x-api.rule=Host(`api-proyecto-x.extension.ucv.ve`)"
      - "traefik.http.routers.proyecto-x-api.entrypoints=web"
      - "traefik.http.services.proyecto-x-api.loadbalancer.server.port=3000"
```

### Base de datos aislada (opcional)

Agregar solo si el proyecto no puede compartir el Postgres existente (diferente versión, propietario externo, etc.). En la mayoría de casos es más simple usar una base de datos distinta dentro del mismo contenedor `db`.

```yaml
  db-proyecto-x:
    image: postgres:16-alpine
    restart: unless-stopped
    volumes:
      - proyecto_x_data:/var/lib/postgresql/data
    environment:
      POSTGRES_USER: ${PROYECTO_X_DB_USER}
      POSTGRES_PASSWORD: ${PROYECTO_X_DB_PASSWORD}
      POSTGRES_DB: ${PROYECTO_X_DB_NAME}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $PROYECTO_X_DB_USER"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - web
```

No olvidar agregar el volumen al final del archivo:

```yaml
volumes:
  postgres_data:
  landing_storage:
  proyecto_x_data:    # agregar aquí
```

### Checklist para cada proyecto nuevo

- [ ] El frontend tiene `traefik.enable=true` con el DNS correcto
- [ ] El backend tiene `traefik.enable=false` (o su propio DNS si es SPA)
- [ ] Ambos están en la red `web`
- [ ] El backend tiene `depends_on` apuntando a su base de datos con `condition: service_healthy`
- [ ] Las variables de entorno del backend están en un `.env` propio del proyecto (o en el `.env` raíz)
- [ ] Si usa DB propia: el volumen está declarado al final del compose
- [ ] Los nombres de router y servicio en los labels de Traefik son únicos en todo el compose

---

**Última actualización**: 2026-04-17
**Versión del documento**: 1.1
**Responsable**: Equipo DEU UCV

