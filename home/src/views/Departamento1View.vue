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
              >DEPARTAMENTO DE RELACIONES INTERINSTITUCIONALES</router-link
            >
          </li>
        </ul>
      </nav>
    </section>
    <section class="section-container">
      <div class="mision-vision">
        <h1 class="titulo">
          DEPARTAMENTO DE RELACIONES INTERINSTITUCIONALES<span class="icon"
            >🚀</span
          >
        </h1>
        <div class="content-container">
          <div class="paragraphs">
            <p>
              Como parte de la reorganización de las funciones y actividades, se
              reubica el Departamento de Eventos, Comunicación y Relaciones
              Institucionales como departamento de apoyo dada su naturaleza
              transversal prestando servicio a toda la Dirección. Así mismo
              cambia su nombre para identificarse apropiadamente con la
              dimensión e importancia de sus funciones.
            </p>
            <p>
              <span class="icon-p"></span>
              Misión y Visión. <br />
              Promover la imagen institucional en diferentes escenarios a nivel
              nacional a través de actividades educativas que permitan
              transformación de la sociedad, mediante la utilización de medios
              comunicacionales y estrategias en la organización de eventos.
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
    ContactForm,
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
            "Gestionar el establecimiento de relaciones con organizaciones públicas, privadas, comunitarias y del tercer sector, dentro de los planes estratégicos de la Dirección.",
        },
        {
          image: require("@/assets/D.png"),
          title: "FUNCIONES",
          subtitle: "",
          description:
            "Elaborar y coordinar planes y programas de comunicación internos y externos para proyectar la imagen de la Dirección de Extensión Universitaria.",
        },
        {
          image: require("@/assets/P.png"),
          title: "CONTACTO",
          subtitle: "",
          description: `
            <div class="contact-info">
              <p><strong>Coordinador:</strong> Juan Pérez</p>
              <p><strong>Teléfono:</strong> <a href="tel:+123456789">+1 234 567 89</a></p>
              <p><strong>Correo:</strong> juanperez@examplecom </p>
            </div>
          `,
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
  },
};
</script>

<style scoped>
/* Estilo general para enlaces */
.contact-link {
  color: #01695b;
  text-decoration: none;
  font-weight: bold;
}

.contact-link:hover {
  text-decoration: underline;
}
@font-face {
  font-family: "museo-sans";
  src: url("../assets/fonts/MuseoSans-100.ttf");
}
* {
  font-family: "museo-sans";
}
.titulo {
  font-size: 30px;
  font-weight: 700;
  line-height: 1em;
  font-family: "museo-sans";
  padding-bottom: 20px;
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
