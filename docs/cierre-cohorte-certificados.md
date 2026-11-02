# Cierre de cohorte y generación de certificados

Este documento describe lo que hace el sistema desde que un proveedor envía una solicitud de cierre de cohorte con sus tres archivos (participantes aprobados, comprobantes y encuesta) hasta que los participantes pueden validar su certificado.

## Diagrama de flujo

```mermaid
flowchart TD
    A([El proveedor envía la solicitud de cierre<br/>con los archivos correctos]) --> B[El sistema lee la planilla de participantes<br/>y valida cada fila]
    B -->|Planilla inválida| B1[/Error 400 con el número de fila<br/>a corregir/]
    B -->|Planilla válida| C{¿Ya hay una solicitud de cierre<br/>pendiente para esta cohorte?}
    C -->|Sí| C1[/Error: solicitud ya pendiente/]
    C -->|No| D[Se crea la solicitud en estado 'en revisión'<br/>y un certificado pendiente por participante,<br/>cada uno con un código de verificación aleatorio]
    D --> E[Los 3 archivos se guardan<br/>en almacenamiento privado]
    E --> F[Un administrador de la DEU revisa la solicitud]
    F -->|Rechaza| F1[Se notifica el motivo por correo<br/>al proveedor]
    F -->|Aprueba| G[En una sola transacción: se cierra la cohorte,<br/>se crea el enlace del ZIP<br/>y se encola el trabajo de certificados]
    G --> H[Se envía al proveedor el correo<br/>de aprobación del cierre]
    G --> I[El proceso en segundo plano toma el trabajo<br/>revisa cada pocos segundos]
    I --> J[Por cada participante, hasta 4 a la vez:<br/>1. genera el PDF con el diseño de la facultad<br/>y un código QR<br/>2. lo guarda en almacenamiento<br/>3. marca el certificado como emitido]
    J -->|Falla| J1[Reintenta con espera creciente, hasta 5 veces.<br/>Si agota los intentos, queda 'fallido'<br/>y un administrador puede reintentarlo]
    J1 --> I
    J --> K[Se arma un ZIP con todos los PDF]
    K --> L[Se envía al proveedor un correo con el enlace al ZIP,<br/>un enlace personal por participante<br/>y la lista en una hoja de cálculo]
    L --> M[El proveedor entrega a cada participante<br/>su certificado]
    M --> N[El participante que aprobó acude a la oficina central<br/>de la DEU para obtener la firma y el sello]
    N --> O([Certificado completo y verificable<br/>con su código QR])
```

## Explicación del flujo

1. **Envío de la solicitud.** El proveedor adjunta tres archivos: la planilla de participantes aprobados (con la plantilla XLSX que ofrece el sistema), los comprobantes y la encuesta. Antes de crear la solicitud, el sistema valida la planilla fila por fila. Si algo está mal, indica el número de fila que se debe corregir.
2. **Revisión.** La solicitud queda "en revisión" y un administrador de la DEU la revisa. Si la rechaza, el proveedor recibe por correo el motivo. Si la aprueba, la cohorte queda cerrada.
3. **Generación de certificados.** Al aprobar, el sistema genera en segundo plano un PDF por cada participante aprobado, con el diseño de la facultad que avaló el curso. Esto no demora la aprobación: el administrador no espera a que terminen los PDF.
4. **Entrega.** Al terminar, el proveedor recibe un correo con todo lo necesario para entregar los certificados.

## Preguntas frecuentes

### Una vez cerrado el curso, ¿cuánto tarda la generación de certificados?

Normalmente unos pocos minutos, que pueden llegar a unos 10 en cohortes muy numerosas. Si hay una falla temporal, el sistema reintenta solo. Para dar una respuesta sencilla al usuario: **hasta 24 horas**. Si pasa ese tiempo sin recibir el correo, debe contactar a la DEU para que reintenten el proceso.

### ¿Dónde puede el proveedor ver los certificados?

En el **correo electrónico** que recibe la persona que envió la solicitud de cierre, con el asunto "Certificados de la cohorte disponibles". Ese correo incluye:

* Un enlace para descargar **todos los certificados en un ZIP**.
* Un **enlace personal por participante**, donde se verifica y descarga su certificado.
* La lista de participantes con su enlace, adjunta en una hoja de cálculo (`participantes.xlsx`).

El sistema no tiene por ahora una pantalla donde el proveedor consulte los certificados después. Por eso conviene que conserve ese correo.

### ¿Recibirán los participantes un enlace a sus correos con su certificado?

**No.** El correo con los enlaces lo recibe solo el proveedor, que se encarga de hacerlos llegar a los participantes. La columna `Correo` de la planilla es opcional y por ahora solo se guarda, sin usarse para enviar nada.

### ¿Qué pasos deben seguir los participantes que aprobaron para validar su certificado?

1. Recibir su certificado (PDF o enlace personal) de parte del proveedor.
2. Acudir a la **oficina central de la DEU** para obtener la **firma y el sello** del certificado.

### ¿Para qué sirve el código QR del certificado?

Lleva a una página pública de verificación de la DEU. Ahí cualquier persona (por ejemplo, un empleador) puede comprobar que el certificado es auténtico. La página muestra:

* El nombre del participante y su cédula parcialmente oculta.
* El curso, su duración, las fechas y la facultad.
* La fecha de emisión.
* Un botón para descargar el PDF.

Si el código no existe o el certificado fue revocado, la página indica que el certificado **no es válido**.
