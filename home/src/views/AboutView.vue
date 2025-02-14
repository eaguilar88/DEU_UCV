<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="sub-header">
      <h1>Sobre Nosotros</h1>
    </section>
    <AppNavbar />
    <section class="link-section">
      <nav class="breadcrumbs">
        <ul>
          <li>
            <router-link to="/"><span>🚀</span></router-link>
          </li>
          <li>
            <router-link to="/departamentos">SOBRE NOSOTROS</router-link>
          </li>
        </ul>
      </nav>
    </section>
    <section class="section-container">
      <div class="mision-vision">
        <h1 class="titulo">
          DIRECCIÓN DE EXTENSIÓN UNIVERSITARIA <span class="icon">🚀 </span>
        </h1>
        <div class="content-container">
          <div class="paragraphs">
            <p>
              Misión. <br /><br />
              Promover la transferencia de conocimientos y de tecnología, la
              acción social y la prestación de servicios para tener un impacto
              positivo en la sociedad, respondiendo a las necesidades del
              entorno y contribuyendo al desarrollo y bienestar de nuestra
              nuestra comunidad interna y externa.
            </p>
            <p>
              <span class="icon-p"></span>
              Vision. <br />
              <br />
              Promover la imagen institucional en diferentes escenarios a nivel
              nacional a través de actividades educativas que permitan
              transformación de la sociedad, mediante la utilización de medios
              comunicacionales, y estrategias en la organización de eventos.
            </p>
            <p>
              <span class="icon-p"></span>
              Valores. <br />
              <br />
              Los valores de la Dirección de Extensión Universitaria se
              fundamentan en el compromiso social, la responsabilidad, la
              innovación, la inclusión, la colaboración y la excelencia. Se
              busca responder a las necesidades del entorno y contribuir al
              desarrollo y bienestar de la comunidad interna y externa mediante
              acciones planificadas, éticas y creativas. Además, se promueve la
              igualdad de oportunidades y el trabajo conjunto con instituciones
              y actores sociales, impulsando la transferencia de conocimientos y
              tecnología para generar un impacto positivo y sostenible.
            </p>
          </div>
        </div>
      </div>
    </section>
    <section class="objetivos">
      <div class="row-about">
        <div class="about-col">
          <div class="about-container-img">
            <img src="../assets/Graduandos.jpg" />
          </div>
        </div>
        <div class="about-col">
          <h1>Funciones</h1>
          <p>
            La Dirección de Extensión Universitaria de la UCV trabaja para
            fortalecer el vínculo entre la universidad y la sociedad a través de
            programas que promueven el desarrollo social, cultural y educativo.
            A continuación, se detallan los objetivos y funciones que guían sus
            acciones en la proyección y el impacto comunitario..
          </p>
          <div class="btn-container">
            <button @click="openDrawer" class="hero-btn">CONOCE MÁS</button>
            <transition
              name="slide"
              @before-enter="beforeEnter"
              @enter="enter"
              @leave="leave"
            >
              <Drawer v-if="isDrawerOpen" @close="closeDrawer" />
            </transition>
          </div>
        </div>
      </div>
    </section>
    <MenuBar :menuItems="menuItems" />
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
              description="La Comisión de Extensión creada en 1988 con la finalidad de dar impulso a las actividades de extensión universitaria y proponer líneas de acción inmediatas que permitieran la integración de todas las Facultades y Dependencias. A partir de 1992 se inició un proceso de reestructuración cuyos resultados se plasmaron el 11 de diciembre de 1995 cuando el Consejo Universitario aprobó la creación de la Coordinación Central de Extensión y el 27 de noviembre de 2002, por decisión del Consejo Universitario, se le otorga rango de Dirección"
              @close="handleClose"
            />
          </div>
        </div>
      </div>
    </section>
    <img src="../assets/diagramafinal.png" class="organigrama" />
  </div>
</template>

<script>
import AppNavbar from "../components/appNavbar";
import Drawer from "../components/Sidebar.vue";
import ContentBar from "../components/content-bar.vue";
import MenuBar from "../components/MenuBar.vue";

export default {
  name: "AboutView",
  components: {
    Drawer,
    AppNavbar,
    ContentBar,
    MenuBar,
  },
  data() {
    return {
      isDrawerOpen: false,
      showContentBar: false,
      menuItems: [
        {
          image: require("@/assets/josephp.jpg"),
          title: "Ing. José Antonio Fernández",
          subtitle: "Sub-Director",
          description:
            "Subdirector de la Dirección de Extensión Universitaria, enfocado en fortalecer las relaciones interinstitucionales y coordinar iniciativas que impulsen la innovación y el desarrollo.",
        },
        {
          image: require("@/assets/mercyphp.jpg"),
          title: "Prof. Mercy Ospina",
          subtitle: "Directora",
          description:
            "Directora de la Dirección de Extensión Universitaria, comprometida con la promoción del conocimiento y la acción social para generar un impacto positivo en la comunidad universitaria y la sociedad en general.",
        },
        {
          image: require("@/assets/secretariaphp.jpg"),
          title: "Sra. Eglin González",
          subtitle: "Secretaria Ejecutiva",
          description:
            "Secretaria General de la Dirección de Extensión Universitaria, responsable de la organización y el soporte administrativo para garantizar el cumplimiento de los objetivos estratégicos del área.",
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
    handleClose() {
      this.showContentBar = false;
    },
  },
};
</script>

<style scoped>
* {
  font-family: museo-sans;
}
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
h1 {
  font-size: 36px;
}

p {
  color: #fff;
  font-size: 14px;
  font-weight: 300;
  line-height: 22px;
  padding: 10px;
}

/*\//*\*/

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
.hero-btn {
  display: inline-block;
  text-decoration: none;
  color: #000000;
  border: 1px solid #000000;
  padding: 12px 34px;
  font-size: 13px;
  background: transparent;
  position: relative;
  cursor: pointer;
  border-radius: 50px;
  width: 200px;
}
.hero-btn:hover {
  border: 1px solid #01695b;
  background: #01695b;
  transition: 1s;
}
.btn-container {
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

/*** */
.layer:hover .menuItems {
  background: #01695b;
}

/** */

.content-container {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  gap: 20px;
  padding-bottom: 70px;
}
.icon {
  margin-right: 5px;
  font-size: 2rem;
}
.section-container {
  width: 100%;
  background-image: linear-gradient(#01695b, #01695bef),
    url("../assets/deuimg.jpg");
}
.titulo {
  font-size: 30px;
  font-weight: 700;
  line-height: 1em;
  font-family: "museo-sans";
  padding-bottom: 20px;
}
.mision-vision {
  padding-top: 80px;
  width: 70%;
  background-image: linear-gradient(#01695b, #01695bef),
    url("../assets/deuimg.jpg");
  color: #fff;
  gap: 20px;
  margin: 0 auto;
  text-align: left;
}

.paragraphs {
  flex: 1;
  flex-direction: column;
}
.paragraphs p {
  margin-bottom: 20px; /* Espacio entre párrafos */
  font-size: 20px;
  font-weight: 300;
}
.image {
  max-width: 300px;
  margin-left: 20px;
  padding-right: 20px;
  padding-bottom: 50px;
}
.organigrama {
  width: 100%;
  height: 100%;
}
/** */
.link-section {
  background-color: #025247;
}
.breadcrumbs {
  font-family: "Arial", sans-serif;
  font-size: 16px;
  display: flex;
  align-items: center;
  padding: 10px 20px;
  background-color: #025247; /* Fondo claro */
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-left: 100px;
}

.breadcrumbs ul {
  list-style: none;
  display: flex;
  gap: 10px; /* Espacio entre los elementos */
  margin: 0;
  padding: 0;
}

.breadcrumbs li {
  display: flex;
  align-items: center;
}

.breadcrumbs li:not(:last-child)::after {
  content: "›"; /* Separador entre enlaces */
  margin-left: 10px;
  margin-right: 10px;
  color: #ffffff; /* Color del separador */
}

.breadcrumbs a {
  text-decoration: none;
  color: #ffffff; /* Azul profesional */
  font-weight: 500; /* Peso medio */
  transition: color 0.3s;
}

.breadcrumbs a:hover {
  color: #ffffff; /* Azul más oscuro en hover */
}

.icon {
  font-size: 20px;
  margin-right: 5px; /* Espacio entre el icono y el texto */
}
</style>
