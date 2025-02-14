<template>
  <div class="menu-global">
    <div class="row">
      <div v-for="(item, index) in menuItems" :key="index" class="menu-col">
        <div class="image-container">
          <img :src="item.image" />
          <div class="layer">
            <div class="text-box">
              <h3>{{ item.title }}</h3>
              <h5>{{ item.subtitle }}</h5>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "MenuBar",
  props: {
    menuItems: {
      type: Array,
      required: true,
    },
  },
};
</script>

<style scoped>
* {
  font-family: museo-sans;
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
  padding: 70px;
  opacity: 1;
  transition: all 0.5s ease;
}

/* Efecto hover sobre el texto */
.layer:hover .text-box h3,
.layer:hover .text-box h5,
.layer:hover .text-box p {
  opacity: 1;
  transform: translateY(0);
}

h3,
h5 {
  margin: 10px 0;
  text-align: center;
  color: #fff;
  transform: translateY(30px);
  transition: all 0.5s ease;
}

p {
  margin: 10px 0;
  text-align: center;
  color: #fff;
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.5s ease;
}

@media (max-width: 700px) {
  .menu-col {
    flex-direction: column;
  }
}
</style>
