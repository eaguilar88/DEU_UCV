<template>
  <div class="container">
    <div class="card-profile" v-for="(profile, index) in profiles" :key="index">
      <div class="profile-image">
        <img :src="profile.image" :alt="profile.name" />
      </div>
      <div class="profile-info">
        <h2 class="profile-name">{{ profile.name }}</h2>
        <p class="profile-description">{{ profile.description }}</p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      profiles: [
        {
          name: "Julie Watson",
          description: "UX / UI developer",
          image: require("../assets/estudiante.jpg"),
        },
        {
          name: "Marc McKnew",
          description: "Front-End developer",
          image: require("../assets/estudiante.jpg"),
        },
        {
          name: "Jenny McKnew",
          description: "Back-End developer",
          image: require("../assets/estudiante.jpg"),
        },
        {
          name: "Jenny McKnew",
          description: "Back-End developer",
          image: require("../assets/estudiante.jpg"),
        },
      ],
    };
  },
};
</script>

<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
  padding: 20px;
  width: 100%;
  max-width: 1200px;
  margin: auto;
}

.card-profile {
  margin-top: 50px;
  display: flex;
  align-items: center;
  background-color: #01695b;
  border-radius: 70px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 450px;
  overflow: hidden;
  flex-direction: row; /* Asegura la dirección horizontal */
}

.profile-image {
  width: 130px;
  height: 130px;
  border-radius: 50%;
  overflow: hidden;
  margin-right: 20px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.profile-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-info {
  flex: 1; /* Asegura que tome todo el espacio disponible */
  display: flex;
  text-align: left;
  flex-direction: column;
}

.profile-name {
  font-size: 1.2em;
  color: #fff;
  font-weight: bold;
  margin-bottom: 8px;
}

.profile-description {
  font-size: 1em;
  color: #fff;
}

@media (max-width: 768px) {
  .card-profile {
    flex-direction: row; /* Mantiene la dirección horizontal en pantallas pequeñas */
    align-items: center;
    text-align: left; /* Alinea el texto a la izquierda */
  }

  .profile-image {
    margin-right: 20px; /* Asegura el margen derecho en pantallas pequeñas */
    margin-bottom: 0; /* Elimina el margen inferior en pantallas pequeñas */
  }
}
</style>
