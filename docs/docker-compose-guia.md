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

## 4. Recursos Adicionales

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

**Última actualización**: 2026-01-30
**Versión del documento**: 1.0
**Mantenedor**: Equipo DEU UCV

