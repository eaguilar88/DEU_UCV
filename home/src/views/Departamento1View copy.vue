<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="sub-header">
      <h1>Departamentos</h1>
    </section>
    <AppNavbar />
    <section class="mision-vision">
      <div class="about-content">
        <h1>PROGRAMAS REGIONALES</h1>
        <p>
          Hacer partícipe a la Universidad Central de Venezuela del desarrollo
          local y regional, coordinando acciones dirigidas a cumplir el objetivo
          de “contribuir a la solución de las diferentes problemáticas socio
          comunitarias de las regiones del país, donde la Universidad Central de
          Venezuela”, promoviendo así la vinculación de la Universidad con las
          comunidades y contribuyendo al desarrollo social, cultural, científico
          y tecnológico del país. y regional. Con esto se busca que el programa
          sea un puente articulador de soluciones para satisfacer las
          necesidades de las comunidades tales como: cursos de actualización,
          mejoramiento profesional y en la medida de las posibilidades, cursos
          de pre y postgrado, asistencia médico-odontológica, desarrollo de
          actividades deportivas y culturales, además de brindar asesorías en
          materia agroindustrial, jurídica, urbanística, de transporte y
          vialidad, entre otras áreas de acción.
        </p>
      </div>
    </section>

    <section>
      <div class="menu-global">
        <div class="row">
          <!-- Columna 1 -->
          <div class="menu-col" v-for="(item, index) in menuItems" :key="index">
            <div class="image-container">
              <img :src="item.image" alt="Imagen del menú" />
              <div class="layer">
                <div class="text-box">
                  <h3>{{ item.title }}</h3>
                  <div class="btn-container">
                    <button
                      @click="openContentBar(item.title, item.description)"
                      class="hero-btn"
                    >
                      CONOCE MÁS
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <ContentBar
        :isVisible="isContentBarVisible"
        :title="currentTitle"
        :description="currentDescription"
        @close="closeContentBar"
      />
    </section>

    <section class="objetivos">
      <div class="row-about">
        <div class="about-col">
          <div class="about-container-img">
            <img src="../assets/Graduandos.jpg" />
          </div>
        </div>
        <div class="about-col">
          <h1>Reseña Historica</h1>
          <p>
            Lorem ipsum dolor sit amet consectetur adipisicing elit. Nostrum sit
            eaque beatae consequatur, repudiandae mollitia blanditiis illo
            quaerat inventore perspiciatis vitae. Nesciunt harum, deleniti atque
            voluptatem rem fugit iure ipsum.
          </p>
          <div class="btn-container">
            <button @click="showContentBar = true" class="hero-btn">
              CONOCE MÁS
            </button>
            <ContentBar
              :isVisible="showContentBar"
              title="RESEÑA HISTORICA"
              description="La Comisión de Extensión creada en 1988 con la finalidad de dar impulso a las actividades de extensión universitaria y proponer líneas de acción inmediatas que permitieran la integración de todas las Facultades y Dependencias. A partir de 1992 se inició un proceso de reestructuración cuyos resultados se plasmaron el 11 de diciembre de 1995 cuando el Consejo Universitario aprobó la creación de la Coordinación Central de Extensión y el 27 de noviembre de 2002, por decisión del Consejo Universitario, se le otorga rango de Dirección."
              @close="handleClose"
            />
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import AppNavbar from "../components/appNavbar";
import ContentBar from "../components/content-bar.vue";

export default {
  name: "Departamento1View",
  components: {
    AppNavbar,
    ContentBar,
  },
  data() {
    return {
      isDrawerOpen: false,
      isContentBarVisible: false,
      showContentBar: false,
      currentTitle: "",
      currentDescription: "",
      menuItems: [
        {
          image: require("@/assets/L.png"),
          title: "OBJETIVOS",
          subtitle: "",
          description:
            "Identificar las necesidades de desarrollo social y económico de las comunidades locales y regionales. Diseñar y ejecutar programas dentro de un marco de acción estratégico para la universidad, emanado desde la Dirección de Extensión Universitaria. Fortalecer el trabajo y la vinculación con las diversas comunidades locales y regionales, pertenecientes a los diferentes sectores.",
        },
        {
          image: require("@/assets/D.png"),
          title: "FUNCIONES",
          subtitle: "",
          description:
            "Realizar estudios y diagnósticos para identificar las necesidades específicas de las comunidades en las que se implementarán los programasDiseñar, ejecutar y evaluar programas.Diseñar y ejecutar programas y proyectos que respondan a las necesidades identificadas, en el marco de las áreas de acción establecidas.Gestionar los recursos humanos, financieros y materiales necesarios para la ejecución de los programas.Monitorear y evaluar el impacto de los programas en las comunidades, con el fin de realizar los ajustes necesarios para mejorar su efectividad.Fortalecer la vinculación con las comunidades a través de la comunicación permanente y la participación activa en la gestión de los programas.Difundir los resultados: Los Programas Regionales deben difundir los resultados de sus investigaciones y evaluaciones. Esto se puede hacer a través de publicaciones, conferencias y talleres.Fortalecer las redes de trabajo: Los Programas Regionales deben fortalecer las redes de trabajo con otras instituciones que trabajan con las comunidades. Esto permitirá compartir experiencias y recursos.Es importante destacar que estas son solo algunas de las posibles funciones que tienen los Programas Regionales. Las funciones específicas de cada programa dependerá de las necesidades de la comunidad, de los convenios suscritos y de los recursos disponibles",
        },
        {
          image: require("@/assets/P.png"),
          title: "CONTACTO",
          subtitle: "",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
        },
      ],
    };
  },
  methods: {
    openDrawer() {
      this.isDrawerOpen = true;
    },
    closeDrawer() {
      this.isDrawerOpen = false;
    },
    openContentBar(title, description) {
      this.currentTitle = title;
      this.currentDescription = description;
      this.isContentBarVisible = true;
    },
    closeContentBar() {
      this.isContentBarVisible = false;
    },
    handleClose() {
      this.showContentBar = false;
    },
    beforeEnter(el) {
      el.style.transform = "translateX(100%)";
    },
    enter(el, done) {
      el.offsetHeight; // Trigger reflow
      el.style.transition = "transform 0.3s ease-in-out";
      el.style.transform = "translateX(0)";
      done();
    },
    leave(el, done) {
      el.style.transition = "transform 0.3s ease-in-out";
      el.style.transform = "translateX(100%)";
      done();
    },
  },
};
</script>
<style scoped>
.sub-header {
  height: 30vh;
  width: 100%;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  background-image: linear-gradient(
      to top,
      rgba(0, 0, 0, 0.9),
      rgba(0, 0, 0, 0.1)
    ),
    url("../assets/graduacion.jpg");
  background-position: center;
  background-size: cover;
  text-align: center;
  color: #fff;
}
.sub-header h1 {
  padding-top: 100px;
  text-align: left;
  padding-left: 140px;
}
@media (max-width: 700px) {
  .sub-header h1 {
    padding-left: 30px;
  }
}
/** */
.about-content {
  width: 80%;
  margin: auto;
  text-align: center;
  padding-top: 10px;
  padding-bottom: 10px;
}
.about-content p {
  color: #fff;
  text-align: left;
}
h1 {
  font-size: 36px;
}
.about-content h1 {
  text-align: left;
  color: white;
  font-weight: 1;
}
p {
  color: #fff;
  font-size: 30px;
  padding: 10px;
}

h2 {
  font-size: 40px;
}

/*\//*\*/
.mision-vision {
  padding-top: 80px;
  padding-bottom: 50px;
  width: 100%;
  margin: auto;
  text-align: center;
  background-image: linear-gradient(#01695b, #01695bef),
    url("../assets/deuimg.jpg");
  color: #fff;
  display: flex;
  flex-direction: column;
}
.mision-vision-header {
  display: flex;
  align-items: center;
  color: white;
}
.about-content p {
  display: flex;
  align-items: center;
  padding-left: 50px;
  color: white;
}
.about-content {
  width: 70%;
}
.icon {
  width: 40px;
  height: 40px;
  margin-right: 10px; /* Espacio entre la imagen y el título */
}

h3 {
  text-align: center;
  font-weight: 100;
  margin: 10px 0;
}
.row-about {
  margin-top: 5%;
  display: flex;
  justify-content: center;
  gap: 50px;
}
@media (max-width: 700px) {
  .row-about {
    flex-direction: column;
  }
}

/** */
.objetivos {
  width: 70%;
  margin: auto;
  padding-top: 50px;
  padding-bottom: 50px;
}
.about-col {
  flex-basis: 48%;
  padding: 30px 2px;
}
.about-col img {
  width: 100%;
}
.about-col p {
  color: black;
  padding: 15px 0 25px;
  text-align: left;
}
.about-col h1 {
  padding-top: 0;
  text-align: left;
}

/** */
.drawer-enter-active,
.drawer-leave-active {
  transition: transform 1s ease, opacity 1s ease;
}
.drawer-enter, .drawer-leave-to /* .drawer-leave-active in <2.1.8 */ {
  transform: translateX(100%);
  opacity: 0;
}
/** */
.menu-content h1 {
  width: 80%;
  margin: auto;
  color: #01695b;
  text-align: left;
  padding-top: 10px;
  padding-bottom: 30px;
}
/** */

/** */

.btn-container {
  display: none; /* Usa flexbox para alinear el contenido */
  justify-content: center; /* Centra horizontalmente */
  align-items: center; /* Centra verticalmente si es necesario */
  position: absolute;
  bottom: 20px; /* Espacio desde el borde inferior de la columna */
  left: 0; /* Alinea el contenedor al borde izquierdo de la columna */
  width: 100%; /* Asegura que el contenedor ocupe todo el ancho de la columna */
  height: auto; /* Ajusta la altura si es necesario */
  text-align: center; /* Centra el texto dentro del contenedor */
}

.menu-col:hover .btn-container {
  display: block; /* Muestra el botón al hacer hover sobre la columna */
}
.hero-btn {
  display: inline-block;
  text-decoration: none;
  color: white;
  border: 1px solid #ffffff;
  padding: 12px 34px;
  font-size: 13px;
  background: transparent;
  cursor: pointer;
  border-radius: 50px;
  width: 200px; /* Ancho del botón */
  /* margin: 0 auto; No es necesario con flexbox en el contenedor */
}
.menu-col {
  position: relative; /* Asegura que el contenedor de imagen tenga posición relativa */
  overflow: hidden;
}

.hero-btn:hover {
  background: #01695b;
  transition: 1s;
}
/** */
/*** */
.row {
  display: flex;
  flex-wrap: wrap;
}

@media (max-width: 700px) {
  .row {
    flex-direction: column;
  }
}

.menu-global {
  width: 100%;
  margin: 0;
  text-align: center;
  padding-top: 0px;
  overflow: hidden;
}

.row .menu-col {
  flex-basis: 33.1%;
  position: relative;
  overflow: hidden;
  padding-left: 0px;
  padding-right: 0px;
}

.image-container {
  position: relative;
  width: 100%;
  height: 75vh;
}

.menu-col img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Capa negra siempre visible */
.layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5); /* Capa negra con 50% de opacidad */
  display: flex;
  justify-content: center;
  align-items: center;
  transition: 0.5s;
}

/* Efecto hover sobre la capa */
.layer:hover {
  background: #01695b; /* Aumenta la opacidad al hacer hover */
}

.text-box {
  color: #fff;
  text-align: center;
  padding: 20px;
  opacity: 1;
  transition: all 0.5s ease;
}

/* Efecto hover sobre el texto */
.layer:hover .text-box h3,
.layer:hover .text-box h5,
.layer:hover .text-box p {
  opacity: 0;
  transform: translateY(0);
}

@media (max-width: 700px) {
  .menu-col {
    flex-direction: column;
  }
}
</style>
