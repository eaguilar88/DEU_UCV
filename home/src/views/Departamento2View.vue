<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="sub-header">
      <h1>Departamentos</h1>
    </section>
    <AppNavbar />
    <section class="link-section">
      <nav class="breadcrumbs">
        <ul>
          <li>
            <router-link to="/"><span>🚀</span></router-link>
          </li>
          <li><router-link to="/departamentos">DEPARTAMENTOS</router-link></li>
          <li>
            <router-link to="/departamentos"
              >DEPARTAMENTO EXTENSIÓN SOCIAL UNIVERSITARIA</router-link
            >
          </li>
        </ul>
      </nav>
    </section>
    <section class="section-container">
      <div class="mision-vision">
        <h1 class="titulo">
          DEPARTAMENTO EXTENSIÓN SOCIAL UNIVERSITARIA<span class="icon"
            >🚀
          </span>
        </h1>
        <div class="content-container">
          <div class="paragraphs">
            <p>
              La propuesta de cambio de nombre del Departamento de Proyectos de
              Extensión y Programas Estudiantiles, a Departamento de Extensión
              Social Universitaria, focaliza la función social de la extensión,
              ampliando la perspectiva de lo meramente asistencial hacia el
              bienestar social, el fortalecimiento del tejido social y la
              proyección de la Extensión Universitaria como función ejecutora
              entre los fondos de acción, inversión y desarrollo social.
            </p>
            <p>
              <span class="icon-p"></span>
              Mision y Vision. <br />
              Diseñar, planificar y coordinar la acción social universitaria a
              través de diferentes actividades articuladas entre estudiantes
              organizados en grupos de extensión y comunidades, fondos
              multilaterales, organizaciones de la sociedad civil e iniciativas
              de responsabilidad social empresarial de organizaciones públicas y
              privadas; sirviendo también de enlace entre las Facultades y
              Escuelas con organizaciones mencionadas. Estas actividades están
              dirigidas a la creación de bienestar social, el fortalecimiento
              del tejido social y la proyección de la Extensión Universitaria.
            </p>
          </div>
        </div>
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
    <ContactForm />
  </div>
</template>

<script>
import AppNavbar from "../components/appNavbar";
import ContentBar from "../components/content-bar.vue";
import ContactForm from "../components/ContactForm.vue";

export default {
  name: "Departamento1View",
  components: {
    AppNavbar,
    ContentBar,
    ContactForm,
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
            "Facilitar el acceso a formaciones no académicas dirigidas a la actualización y el mejoramiento profesional de toda la comunidad universitaria. Mantener una oferta educativa y accesible a un público amplio que denota el compromiso social de la Universidad por la sociedad.",
        },
        {
          image: require("@/assets/D.png"),
          title: "FUNCIONES",
          subtitle: "",
          description:
            "Diseñar, elaborar y coordinar programas y proyectos que tengan que ver con la proyección extensionista universitaria.Desarrollar un sistema de información para el registro, monitoreo y seguimiento de proyectos y programas de extensión universitaria a nivel nacional.Conceptuar y proponer nuevos productos o servicios que posibiliten la generación de recursos propíos, a través de planes de funcionamiento como patrocinantes, convenios marco y específicos entre instituciones públicas y privadas.Participar en el desarrollo de la visión estratégica de la Dirección de Extensión Universitaria, a través de planes estratégicos que permitan demostrar la factibilidad de la permanencia y sustentabilidad de los productos ofertados.Controlar y evaluar los proyectos y programas que se realicen a través de convenios.Cooperar en la oferta de programas de inducción a estudiantes y profesores como actividad preparatoria al inicio de las prácticas comunitarias.Identificar las necesidades más apremiantes en las comunidades, con la finalidad de disponer de una base de datos de problemas comunitarios, la que junto a las que dispongan las Escuelas y Facultades de la UCV, pueden ser considerados como insumos para determinar acciones conjuntas.Diseñar propuestas para la firma de Rector(a), convenios para la presentación del servicio comunitario, con los Consejos Locales de Planificación Pública, Consejos Estadales de Planificación y Coordinación de Políticas Públicas, instituciones y organizaciones públicas o privadas y comunidad organizada entre otras.Evaluar los proyectos presentados por los sectores con iniciativa, a objeto de ser considerada su aprobación.Garantizar mecanismos de información para las distintas comunidades, sobre las áreas que competen a la Universidad Central de Venezuela relativas al Servicio Comunitario.Las demás funciones que le confiere las leyes y reglamentos, normas y su supervisor inmediato",
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
@font-face {
  font-family: "museo-sans";
  src: url("../assets/fonts/MuseoSans-100.ttf");
}
* {
  font-family: "museo-sans";
}
.titulo {
  letter-spacing: -2px;
  font-size: 35px;
  font-weight: 700;
  line-height: 1em;
  font-family: "museo-sans";
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

p {
  color: #fff;
  font-size: 30px;
  padding: 10px;
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

.drawer-enter-active,
.drawer-leave-active {
  transition: transform 1s ease, opacity 1s ease;
}
.drawer-enter, .drawer-leave-to /* .drawer-leave-active in <2.1.8 */ {
  transform: translateX(100%);
  opacity: 0;
}
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

.content-container {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  gap: 20px;
  padding-bottom: 70px;
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

.icon {
  margin-right: 5px;
  font-size: 2rem;
}
.section-container {
  width: 100%;
  background-image: linear-gradient(#01695b, #01695bef),
    url("../assets/deuimg.jpg");
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
