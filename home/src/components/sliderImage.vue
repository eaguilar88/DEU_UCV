<template>
  <div class="container-fluid p-0">
    <div class="row no-gutters">
      <div class="col-12">
        <div
          id="carouselExampleIndicators"
          class="carousel slide"
          data-bs-ride="carousel"
          style="height: 75vh; overflow: hidden"
        >
          <div class="carousel-indicators">
            <button
              v-for="(slide, index) in slides"
              :key="index"
              type="button"
              :data-bs-target="'#carouselExampleIndicators'"
              :data-bs-slide-to="index"
              :class="{ active: index === 0 }"
              :aria-label="'Slide ' + (index + 1)"
            ></button>
          </div>
          <div class="carousel-inner">
            <div
              v-for="(slide, index) in slides"
              :key="index"
              :class="['carousel-item', { active: index === 0 }]"
            >
              <img :src="slide.image" class="d-block w-100" alt="..." />
              <div class="carousel-caption">
                <p class="date">{{ slide.date }}</p>
                <h1>{{ slide.title }}</h1>
                <p>{{ slide.text }}</p>
              </div>
            </div>
          </div>
          <button
            class="carousel-control-prev"
            type="button"
            data-bs-target="#carouselExampleIndicators"
            data-bs-slide="prev"
          >
            <span class="carousel-control-prev-icon" aria-hidden="true"></span>
            <span class="visually-hidden">Previous</span>
          </button>
          <button
            class="carousel-control-next"
            type="button"
            data-bs-target="#carouselExampleIndicators"
            data-bs-slide="next"
          >
            <span class="carousel-control-next-icon" aria-hidden="true"></span>
            <span class="visually-hidden">Next</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap/dist/js/bootstrap.min.js";
import * as Bootstrap from "bootstrap";

export default {
  data() {
    return {
      slides: [
        {
          date: "12/08/2023",
          title: "DEU llevó a cabo 6ta reunión de Coordinadores de Extensión",
          text: "Texto pequeño para el slider 1",
          image: require("@/assets/nubes.png"),
        },
        {
          date: "13/08/2023",
          title: "DEU llevó a cabo 6ta reunión de Coordinadores de Extensión",
          text: "Texto pequeño para el slider 2",
          image: require("@/assets/nubes.png"),
        },
        {
          date: "14/08/2023",
          title: "DEU llevó a cabo 6ta reunión de Coordinadores de Extensión",
          text: "Texto pequeño para el slider 3",
          image: require("@/assets/nubes.png"),
        },
      ],
    };
  },
  mounted() {
    var myCarousel = document.querySelector("#carouselExampleIndicators");
    new Bootstrap.Carousel(myCarousel);
  },
};
</script>

<style scoped>
.carousel-inner {
  height: 75vh;
}

.carousel-item img {
  height: 75vh;
  object-fit: cover;
}

.carousel-caption {
  position: absolute;
  bottom: 5%;
  left: 10%;
  text-align: left;
  color: white;
  border-radius: 10px;
}

.carousel-caption .date {
  color: rgba(255, 230, 9, 0.8);
}

.carousel-caption h1 {
  font-size: 40px;
  text-shadow: 2px 7px 4px rgba(0, 0, 0, 0.8);
}
</style>
