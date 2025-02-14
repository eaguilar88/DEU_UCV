<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="header">
      <div class="text-box">
        <h1>Dirección de Extensión <br />Universitaria</h1>
        <div class="headerbtn-container">
          <router-link to="/about" class="hero-btn"
            >CONOCE MÁS SOBRE NOSOTROS</router-link
          >
        </div>
      </div>
    </section>
    <MenuGlobal :menuItems="menuItems" />
    <DiploCont />
    <sliderImage />
    <section class="noticias">
      <router-link to="/noticias" class="news-link">
        <h2>Explora más noticias y anuncios</h2>
      </router-link>
      <Gallery :items="homeItems" />
    </section>
  </div>
</template>

<script>
import sliderImage from "../components/sliderImage.vue";
import Gallery from "../components/GalleryView.vue";
import MenuGlobal from "../components/MenuGlobal.vue";
import DiploCont from "../components/DiplomadosCont.vue";

export default {
  name: "HomeView",
  components: {
    sliderImage,
    Gallery,
    MenuGlobal,
    DiploCont,
  },
  data() {
    return {
      homeItems: [
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title:
            "DEU y la Jefatura de Protocolo dictaron Taller de Unificación de Criterios Protocolares",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "DEU y aliados se unen al Proyecto de Tacarigua",
        },
      ],
      menuItems: [
        {
          image: require("@/assets/L.png"),
          icon: require("@/assets/lista.png"),
          title: "ESPACIOS UNIVERSITARIOS",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
          route: "/espacios-universitarios", // Nueva propiedad para definir la ruta
        },
        {
          image: require("@/assets/D.png"),
          icon: require("@/assets/organizacion.png"),
          title: "ESTRUCTURA ORGANIZATIVA",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
          route: "/departamentos1", // Nueva propiedad para definir la ruta
        },
        {
          image: require("@/assets/P.png"),
          icon: require("@/assets/proyecto.png"),
          title: "CERTIFICACIONES Y AVALES",
          description:
            "Lorem ipsum dolor sit amet consectetur, adipisicing elit. Veritatis, soluta hic eaque natus asperiores eligendi enim provident laborum.",
          route: "/certificaciones-avales", // Nueva propiedad para definir la ruta
        },
      ],
    };
  },
};
</script>
<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  font-family: var(--fontSansSerif);
  font-weight: 400;
  font-size: 13.5px;
  font-family: "rockwell";
}
.header {
  height: 90vh;
  width: 100%;
  background-image: linear-gradient(rgba(0, 0, 0, 0.555), rgba(0, 0, 0, 0.555)),
    url("../assets/biblioteca.jpg");
  background-size: cover;
  background-position: center;
  position: relative;
}

.text-box {
  width: 90%;
  color: #fff;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.text-box h1 {
  font-size: 50px;
  text-shadow: 2px 7px 4px rgba(0, 0, 0, 0.8);
  padding-bottom: 40px;
}

.text-box p {
  margin: 10px 0 40px;
  font-size: 14px;
  color: #fff;
  text-shadow: 2px 7px 4px rgba(0, 0, 0, 0.8);
}

.header .text-box p {
  text-align: left;
}

@media (max-width: 700px) {
  .text-box h1 {
    font-size: 20px;
  }
}

/* */

/*** */
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
}
.hero-btn:hover {
  border: 1px solid #01695b;
  background: #01695b;
  transition: 1s;
}
.btn-container {
  text-align: left;
}
.headerbtn-container .hero-btn {
  color: #fff;
  border: 1px solid #fff;
  text-align: left;
}
/** */
.noticias {
  height: 100vh;
  width: 100%;
  text-align: left;
  padding-top: 50px;
  color: #000000;
}
.texto {
  width: 70%;
}
.noticias h2 {
  color: #000000;
  padding-left: 200px;
}
</style>
