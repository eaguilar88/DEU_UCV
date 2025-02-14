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
              >DIVISIÓN DE PROYECTOS Y PROGRAMAS</router-link
            >
          </li>
        </ul>
      </nav>
    </section>
    <section class="section-container">
      <div class="mision-vision">
        <h1 class="titulo">
          DIVISIÓN DE PROYECTOS Y PROGRAMAS<span class="icon">🚀 </span>
        </h1>
        <div class="content-container">
          <div class="paragraphs">
            <p>
              Los Programas Regionales amplían, a otros espacios geográficos, la
              acción de la extensión universitaria, facilitando además las bases
              para ampliar la investigación y la docencia. Es por esto que los
              Programas Regionales son focos de desarrollo local. <br />
              <br />
              coordinación de los programas locales como el Programa Nva.
              Esparta, Miranda, Amazonas y Delta se constituyen, en esta
              propuesta, en Departamentos que soportan la debida estructura
              funcional y le dan peso institucional a las labores de
              relacionamiento local con entes y organizaciones públicas y
              privadas, especialmente en aquellos programas que cuentan con sede
              propia. <br />
              <br />
              Al enlazar los Programas a través de una División se busca
              centralizar la planificación, ejecución y seguimiento de las
              actividades dentro de una misma estrategia. <br />
              <br />
              En esta División se encuentra también el Departamento de
              Proyectos, Productos y Servicios UCV cuya función es identificar,
              apoyar, visibilizar y conectar proyectos, productos y servicios
              universitarios con la sociedad a través de diferentes mecanismos
              como spin-off, startups, la incubación, la aceleración, el
              licenciamiento y otros conformados por la empresa y la
              universidad, que serán ejecutados por los entes universitarios
              destinados a ese fin.
            </p>
            <p>
              <span class="icon-p"></span>
              Mision y Vision. <br />
              <br />
              Coordinar los Programas Extensión Universitaria en las localidades
              y regiones, y los Proyectos Universitarios para conectar con la
              sociedad, convirtiéndose en actores del desarrollo regional y
              local, ampliando la cobertura y el impacto de las acciones de la
              Universidad a diferentes espacios de la geografía nacional
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
            "Planificación, diseño, ejecución y evaluación de programas y proyectos de extensión universitaria en las localidades y regiones del país. Promover el establecimiento de alianzas estratégicas con actores públicos, privados y sociales para la ejecución de programas y proyectos.",
        },
        {
          image: require("@/assets/D.png"),
          title: "FUNCIONES",
          subtitle: "",
          description:
            "Formular y evaluar la factibilidad de realizar programas y proyectos de orientación educativa.Asegurar el cumplimiento de los lineamientos metodológicos requeridos.Realizar informes de avance de los programas y proyectos de extensión educativa.Brindar información de los programas y proyectos de extensión educativa.Participar en la organización de cursos y talleres, así como en la coordinación de estos con las Facultades y Dependencias Centrales de la Universidad que lo requieran.Elaborar la planificación de cursos de capacitación, presenciales o a distancia a ser dictados por la Dirección de Extensión.Coordinar la divulgación de todos los programas educativos de extensión de la Universidad.Las demás funciones que le confieren las leyes y reglamentos, normas y su supervisor inmediato",
        },
        {
          image: require("@/assets/P.png"),
          title: "CONTACTO",
          subtitle: "",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum",
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
