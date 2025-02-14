<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="sub-header">
      <h1>Departamentos</h1>
    </section>
    <AppNavbar />
    <section class="mision-vision">
      <div class="about-content">
        <h1>PROGRAMA NUEVA ESPARTA</h1>
        <p>
          Lorem ipsum dolor sit amet consectetur adipisicing elit. Dolor optio
          neque delectus obcaecati, maiores alias aperiam, magni quis deserunt
          illum quam. Dignissimos sapiente similique voluptatibus! Temporibus
          earum dolorem veniam pariatur. Lorem ipsum dolor sit amet consectetur
          adipisicing elit. Dolor optio neque delectus obcaecati, maiores alias
          aperiam, magni quis deserunt illum quam. Dignissimos sapiente
          similique voluptatibus! Temporibus earum dolorem veniam pariatur.
        </p>
      </div>
      <div class="about-content">
        <div class="mision-vision-header">
          <img src="../assets/comprobado.png" alt="Icono Misión" class="icon" />
          <h2>Misión</h2>
        </div>
        <p>
          Lorem ipsum dolor sit amet consectetur adipisicing elit. Dolor optio
          neque delectus obcaecati, maiores alias aperiam, magni quis deserunt
          illum quam. Dignissimos sapiente similique voluptatibus! Temporibus
          earum dolorem veniam pariatur. Lorem ipsum dolor sit amet consectetur
          adipisicing elit. Dolor optio neque delectus obcaecati, maiores alias
          aperiam, magni quis deserunt illum quam. Dignissimos sapiente
          similique voluptatibus! Temporibus earum dolorem veniam pariatur.
        </p>
      </div>
      <div class="about-content">
        <div class="mision-vision-header">
          <img src="../assets/comprobado.png" alt="Icono Visión" class="icon" />
          <h2>Visión</h2>
        </div>
        <p>
          Lorem ipsum dolor sit amet consectetur adipisicing elit. Dolor optio
          neque delectus obcaecati, maiores alias aperiam, magni quis deserunt
          illum quam. Dignissimos sapiente similique voluptatibus! Temporibus
          earum dolorem veniam pariatur. Lorem ipsum dolor sit amet consectetur
          adipisicing elit. Dolor optio neque delectus obcaecati, maiores alias
          aperiam, magni quis deserunt illum quam. Dignissimos sapiente
          similique voluptatibus! Temporibus earum dolorem veniam pariatur.
        </p>
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
          <h1>Objetivos y Funciones</h1>
          <p>
            Lorem ipsum dolor sit amet consectetur adipisicing elit. Nostrum sit
            eaque beatae consequatur, repudiandae mollitia blanditiis illo
            quaerat inventore perspiciatis vitae. Nesciunt harum, deleniti atque
            voluptatem rem fugit iure ipsum.
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

    <section>
      <div class="menu-global">
        <div class="row">
          <!-- Columna 1 -->
          <div class="menu-col">
            <div class="image-container">
              <img src="../assets/1.jpg" />
              <div class="layer">
                <div class="text-box">
                  <h3>Título 1</h3>
                  <div class="btn-container">
                    <button
                      @click="
                        openContentBar(
                          'Título 1 para ContentBar',
                          'Descripción detallada para ContentBar 1'
                        )
                      "
                      class="hero-btn"
                    >
                      CONOCE MÁS
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Columna 2 -->
          <div class="menu-col">
            <div class="image-container">
              <img src="../assets/2.jpg" />
              <div class="layer">
                <div class="text-box">
                  <h3>Título 2</h3>
                  <div class="btn-container">
                    <button
                      @click="
                        openContentBar(
                          'Título 2 para ContentBar',
                          'Descripción detallada para ContentBar 2'
                        )
                      "
                      class="hero-btn"
                    >
                      CONOCE MÁS
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Columna 3 -->
          <div class="menu-col">
            <div class="image-container">
              <img src="../assets/3.jpg" />
              <div class="layer">
                <div class="text-box">
                  <h3>Título 3</h3>
                  <div class="btn-container">
                    <button
                      @click="
                        openContentBar(
                          'Título 3 para ContentBar',
                          'Descripción detallada para ContentBar 3'
                        )
                      "
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
import Drawer from "../components/Sidebar.vue";
import ContentBar from "../components/content-bar.vue";

export default {
  name: "Departamento1View",
  components: {
    Drawer,
    AppNavbar,
    ContentBar,
  },
  data() {
    return {
      isDrawerOpen: false,
      showContentBar: false,
      menuItems: [
        {
          image: require("@/assets/L.png"),
          title: "PROGRAMAS",
          subtitle: "Directora Ejecutiva",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
        },
        {
          image: require("@/assets/D.png"),
          title: "DEPARTAMENTOS",
          subtitle: "Directora Ejecutiva",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
        },
        {
          image: require("@/assets/P.png"),
          title: "PROYECTOS",
          subtitle: "Directora Ejecutiva",
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
  font-size: 14px;
  font-weight: 300;
  line-height: 22px;
  padding: 10px;
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
