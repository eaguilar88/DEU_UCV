# Arquitectura del Sistema DEU UCV

## Diagrama de Arquitectura

```mermaid
C4Context
    title Sistema de Gestión DEU UCV - Diagrama de Arquitectura

    Enterprise_Boundary(b0, "Sistema DEU UCV") {
        System_Boundary(b1, "Backend (Go)") {
            Container(api, "API Server", "Go/chi", "Maneja las solicitudes HTTP y expone los endpoints REST")
            
            Container(auth, "Servicio de Autenticación", "Go", "Gestiona autenticación y autorización")
            Container(providers, "Servicio de Proveedores", "Go", "Gestiona proveedores y sus documentos")
            Container(courses, "Servicio de Cursos", "Go", "Gestiona cursos y ciclos")
            Container(groups, "Servicio de Grupos", "Go", "Gestiona grupos de extensión")
            Container(storage, "Servicio de Almacenamiento", "Go", "Gestiona archivos usando B2")

            ContainerDb(db, "Base de Datos", "PostgreSQL", "Almacena toda la información del sistema")
            ContainerDb(b2, "Almacenamiento de Archivos", "Backblaze B2", "Almacena documentos y archivos")
        }

        System_Boundary(b2, "Frontend") {
            Container(landing, "Landing Page", "Vue.js", "Sitio web público")
            Container(diplomados, "Portal de Diplomados", "Next.js", "Portal para gestión de cursos")
            Container(grupos, "Portal de Grupos", "Next.js", "Portal para gestión de grupos")
        }

        BiRel(api, auth, "Usa")
        BiRel(api, providers, "Usa")
        BiRel(api, courses, "Usa")
        BiRel(api, groups, "Usa")
        
        Rel(providers, storage, "Usa")
        Rel(courses, storage, "Usa")
        Rel(groups, storage, "Usa")
        
        BiRel(auth, db, "Lee/Escribe")
        BiRel(providers, db, "Lee/Escribe")
        BiRel(courses, db, "Lee/Escribe")
        BiRel(groups, db, "Lee/Escribe")
        
        Rel(storage, b2, "Almacena")

        BiRel(landing, api, "HTTP/JSON")
        BiRel(diplomados, api, "HTTP/JSON")
        BiRel(grupos, api, "HTTP/JSON")
    }

    Enterprise_Boundary(b3, "Sistemas Externos") {
        System_Ext(email, "Servicio de Email", "SMTP")
    }

    Rel(api, email, "Envía emails")

    UpdateElementStyle(auth, $bgColor="lightblue")
    UpdateElementStyle(providers, $bgColor="lightgreen")
    UpdateElementStyle(courses, $bgColor="lightgreen")
    UpdateElementStyle(groups, $bgColor="lightgreen")
    UpdateElementStyle(storage, $bgColor="lightyellow")

```

## Descripción de Componentes

### Backend (Go)
- **API Server**: Punto de entrada principal que maneja todas las solicitudes HTTP y expone los endpoints REST.
- **Servicio de Autenticación**: Gestiona la autenticación y autorización de usuarios.
- **Servicio de Proveedores**: Maneja la lógica de negocio relacionada con proveedores y sus documentos.
- **Servicio de Cursos**: Gestiona la lógica de negocio relacionada con cursos y ciclos.
- **Servicio de Grupos**: Maneja la lógica de negocio relacionada con grupos de extensión.
- **Servicio de Almacenamiento**: Interfaz unificada para el manejo de archivos usando Backblaze B2.

### Frontend
- **Landing Page**: Sitio web público implementado en Vue.js.
- **Portal de Diplomados**: Aplicación Next.js para la gestión de cursos.
- **Portal de Grupos**: Aplicación Next.js para la gestión de grupos de extensión.

### Almacenamiento
- **PostgreSQL**: Base de datos principal que almacena toda la información del sistema.
- **Backblaze B2**: Servicio de almacenamiento en la nube para documentos y archivos.

### Sistemas Externos
- **Servicio de Email**: Sistema externo para el envío de correos electrónicos.

## Flujo de Datos

1. Los clientes (frontend) interactúan con el sistema a través de la API REST.
2. El API Server autentica las solicitudes y las direcciona al servicio correspondiente.
3. Los servicios procesan las solicitudes, interactuando con la base de datos y el almacenamiento según sea necesario.
4. Los archivos se almacenan en Backblaze B2 a través del servicio de almacenamiento.
5. Las notificaciones se envían a través del servicio de email externo.