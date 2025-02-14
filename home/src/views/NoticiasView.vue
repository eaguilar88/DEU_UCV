<template>
  <div style="min-height: 100vh; width: 100%">
    <section class="sub-header">
      <h1>Departamentos</h1>
    </section>
    <AppNavbar />
    <sliderImage />
    <Gallery :items="homeItems" />
  </div>
</template>

<script>
import AppNavbar from "../components/appNavbar";
import sliderImage from "../components/sliderImage.vue";
import Gallery from "../components/GalleryView.vue";

export default {
  name: "HomeView",
  components: {
    sliderImage,
    Gallery,
    AppNavbar,
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
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
        {
          image: require("../assets/mercy3.jpeg"),
          title: "Venezuela y la UCV son campeones del Pre-CMUDE Bogotá 2024",
        },
      ],
    };
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
</style>
