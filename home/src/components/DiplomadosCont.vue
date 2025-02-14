<template>
  <div class="diplomados">
    <div class="row-diplo">
      <div class="text-content">
        <h1>{{ title }}</h1>
        <p>{{ description }}</p>
        <div class="btn-container">
          <a :href="buttonLink" class="hero-btn">{{ buttonText }}</a>
        </div>
      </div>
      <div class="cards-container">
        <div
          class="diplo-col"
          v-for="(diplomado, index) in diplomados"
          :key="index"
        >
          <img :src="diplomado.image" :alt="diplomado.alt" />
          <h3>{{ diplomado.title }}</h3>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "DiploCont",
  data() {
    return {
      title: "Título de los Diplomados",
      description: `Lorem ipsum dolor sit amet consectetur adipisicing elit.
                    Magnam tempora culpa voluptatum illo incidunt porro amet 
                    minima vel illum, illum, eius non possimus placeat 
                    inventore dicta numquam.`,
      buttonText: "EXPLORAR",
      buttonLink: "#",
      diplomados: [
        {
          image: require("../assets/marketing.png"),
          alt: "Descripción de la imagen de marketing",
          title: "MARKETING DIGITAL Y RRSS",
        },
        {
          image: require("../assets/medicine.jpeg"),
          alt: "Descripción de la imagen de medicina",
          title: "ECOGRAFIA PULMONAR",
        },
        {
          image: require("../assets/programer.jpeg"),
          alt: "Descripción de la imagen de programación",
          title: "COMPUTACION DE ALTO RENDIMIENTO",
        },
      ],
    };
  },
};
</script>

<style scoped>
@font-face {
  font-family: "Brandon";
  src: url("../assets/fonts/HvDTrial_Brandon_Grotesque_regular-BF64a625c9311e1.otf");
}
.diplomados {
  width: 80%;
  margin: auto;
  text-align: center;
  padding-top: 50px;
  padding-bottom: 100px;
}

.row-diplo {
  margin-top: 5%;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.text-content {
  flex-basis: 30%;
  text-align: left;
}

.cards-container {
  flex-basis: 65%;
  display: flex;
  flex-wrap: wrap;
  gap: 2%;
}

.diplo-col {
  position: relative;
  flex-basis: 30%;
  text-align: left;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3); /* Sombra en las tarjetas */
  border-radius: 8px; /* Opcional: bordes redondeados para las tarjetas */
  transition: box-shadow 0.3s ease; /* Transición suave en hover */
}
.diplo-col:hover {
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.4); /* Sombra más intensa al pasar el mouse */
}
.diplo-col img {
  width: 100%;
  height: 60vh;
  object-fit: cover;
}

.diplomados p {
  color: black;
}

.diplo-col h3 {
  font-size: 15px;
  text-align: left;
  position: absolute;
  padding-left: 10px;
  padding-right: 10px;
  top: 80%;
  color: #fff;
  font-family: Brandon;
}

@media (max-width: 700px) {
  .row-diplo {
    flex-direction: column;
  }
  .text-content,
  .cards-container {
    flex-basis: 100%;
  }
  .diplo-col {
    flex-basis: 100%;
  }
}
</style>
