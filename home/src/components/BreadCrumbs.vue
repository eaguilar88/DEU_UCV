<template>
  <nav class="breadcrumbs">
    <ul>
      <li>
        <router-link to="/"><span class="icon">🚀</span></router-link>
      </li>
      <li><router-link to="/departamentos">Departamentos</router-link></li>
      <li>
        <router-link to="/departamentos"
          >Departamento de Relaciones Interinstitucionales</router-link
        >
      </li>
    </ul>
  </nav>
</template>

<style scoped>
@font-face {
  font-family: "museo-sans";
  src: url("../assets/fonts/MuseoSans-100.ttf");
}
* {
  font-family: "museo-sans";
}
.breadcrumbs {
  padding: 10px;
  background-color: #025247;
}

.breadcrumbs ul {
  display: flex;
  list-style: none;
  padding: 0;
  padding-left: 100px;
}

/* Eliminar la decoración y color de los enlaces */
.breadcrumbs ul li a {
  text-decoration: none;
  font-size: medium;
  color: white;
}

.breadcrumbs ul li {
  margin-right: 10px;
}

.breadcrumbs ul li::after {
  content: ">";
  margin-left: 10px;
  color: white;
}

.breadcrumbs ul li:last-child::after {
  content: ""; /* Elimina la flecha en el último elemento */
}
</style>
