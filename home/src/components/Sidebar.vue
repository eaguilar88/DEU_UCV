<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="drawer">
    <button class="close-button" @click="$emit('close')">X</button>
    <h2>FUNCIONES</h2>
    <div class="content">
      <div class="column">
        <!-- Items estáticos de la primera columna -->
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Coordinar, ejecutar, organizar y asesorar en materia de
            extensión.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span>Presidir el Consejo Central de Extensión.</span>
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Coordinar, supervisar y evaluar las actividades de extensión de la
            Universidad.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Mantener informado al Consejo Central de Extensión de los hechos
            relevantes del acontecer nacional e internacional que requieran la
            participación de la UCV.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Representar a la Universidad en el Núcleo de Directores de
            Extensión del Consejo Nacional de Universidades.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Representar a la DEU en los Consejos Centrales de las Dependencias
            Centrales, donde los hubiere.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Desarrollar el plan estratégico de actividades de extensión que
            responda a la solución de la problemática nacional.</span
          >
        </div>
      </div>
      <div class="column">
        <!-- Items estáticos de la segunda columna -->
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Evaluar la factibilidad técnica y económica de los proyectos
            presentados a la Dirección.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Elaborar y controlar la ejecución del presupuesto asignado.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Establecer y promover la interacción con personalidades e
            instancias del Estado, Gobiernos Regionales, Locales, Universidades
            Nacionales, Gremios, Asociaciones, Organismos y Empresas vinculadas
            a la labor extensionista.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Promover y desarrollar Convenios de Cooperación Interinstitucional
            nacionales e internacionales con el propósito de facilitar la
            ejecución de programas, proyectos y otras actividades de
            extensión.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Mantener relación directa con los distintos entes de la
            Universidad, a fin de garantizar la presencia e imagen de la DEU en
            la UCV.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Participar activamente en la promoción y divulgación de los logros,
            planes y proyectos de extensión de la UCV.</span
          >
        </div>
        <div class="item">
          <img src="@/assets/ajustes.png" alt="icon" />
          <span
            >Cumplir con las funciones que le sean señaladas en los Reglamentos
            y Normativas emanados del Consejo Universitario y por el
            Rector.</span
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  // No se necesita data ni script en este caso
};
</script>

<style>
.drawer {
  position: fixed;
  top: 0;
  right: 0;
  width: 80%;
  height: 100%;
  background: #faf9f5;
  box-shadow: -2px 0 5px rgba(0, 0, 0, 0.5);
  padding: 20px;
  overflow-y: auto;
  transition: transform 0.3s ease-in-out;
  transform: translateX(100%);
  z-index: 1000;
}

.close-button {
  position: absolute;
  top: 10px;
  left: 10px;
  font-size: 30px;
  border: none;
  cursor: pointer;
  background-color: #01695b;
  width: 50px;
  height: 50px;
  color: white;
  font-family: sans-serif;
}

.drawer h2 {
  margin-top: 40px;
  padding-top: 40px;
  margin-bottom: 20px;
  padding-bottom: 30px;
  padding-left: 40px;
  border-bottom: 2px solid #000000;
}

.content {
  display: flex;
  justify-content: flex-start;
  gap: 20px;
  padding-left: 40px;
  padding-right: 40px;
}

.column {
  display: flex;
  flex-direction: column; /* Alinea los items verticalmente */
  flex: 1;
  max-width: 80%;
}

.item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.item img {
  width: 20px;
  height: 20px;
  margin-right: 10px;
}

@media (max-width: 768px) {
  .drawer {
    width: 100%;
  }
  .column {
    flex: 1 1 100%;
    margin-right: 0;
  }
  .content {
    flex-direction: column; /* Apila las columnas verticalmente en pantallas pequeñas */
    align-items: flex-start; /* Alinea las columnas a la izquierda en pantallas pequeñas */
  }
}
</style>
