<template>
  <nav
    class="navbar navbar-expand-lg navbar-dark"
    style="background-color: #000000"
  >
    <div class="container-fluid">
      <a class="navbar-brand" href="#">Navbar</a>
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarNavDropdown"
        aria-controls="navbarNavDropdown"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div
        class="collapse navbar-collapse justify-content-center"
        id="navbarNavDropdown"
      >
        <ul class="navbar-nav">
          <li class="nav-item">
            <a class="nav-link active" aria-current="page" href="#">HOME</a>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/about">NOSOTROS</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/noticias">NOTICIAS</router-link>
          </li>
          <li class="nav-item dropdown">
            <a
              class="nav-link dropdown-toggle"
              href="#"
              id="navbarDropdownMenuLink3"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              PROGRAMAS
            </a>
            <ul class="dropdown-menu" aria-labelledby="navbarDropdownMenuLink3">
              <li>
                <router-link class="dropdown-item" to="/programas1"
                  >DIVISIÓN DE PROYECTOS Y PROGRAMAS</router-link
                >
              </li>
              <li>
                <router-link class="dropdown-item" to="/programas2"
                  >PROGRAMA MIRANDA</router-link
                >
              </li>
              <li>
                <router-link class="dropdown-item" to="/programas3"
                  >PROGRAMA NUEVA ESPARTA</router-link
                >
              </li>
            </ul>
          </li>
          <li class="nav-item dropdown">
            <a
              class="nav-link dropdown-toggle"
              href="#"
              id="navbarDropdownMenuLink3"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              DEPARTAMENTOS
            </a>
            <ul class="dropdown-menu" aria-labelledby="navbarDropdownMenuLink3">
              <li>
                <router-link class="dropdown-item" to="/departamentos1"
                  >DEPARTAMENTO RELACIONES INTERINSTITUCIONALES</router-link
                >
              </li>
              <li>
                <router-link class="dropdown-item" to="/departamentos2"
                  >DEPARTAMENTO EXTENSIÓN SOCIAL UNIVERSITARIA</router-link
                >
              </li>
              <li>
                <router-link class="dropdown-item" to="/departamentos3"
                  >DEPARTAMENTO EDUCACIÓN CONTINUA Y PERMANENTE</router-link
                >
              </li>
            </ul>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/noticias"
              >GRUPOS DE EXTENSIÓN</router-link
            >
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/noticias">ESPACIOS</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/noticias"
              >CERTIFICACIONES</router-link
            >
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script>
export default {
  name: "AppNavbar",
};
</script>

<style scoped>
@font-face {
  font-family: "Brandon";
  src: url("../assets/fonts/HvDTrial_Brandon_Grotesque_regular-BF64a625c9311e1.otf");
}
* {
  font-family: Brandon;
}
/* Espaciado entre los ítems de la barra de navegación */
.navbar-nav .nav-item {
  margin-right: 40px; /* Ajusta este valor según el espacio deseado */
}

/* Espaciado entre los ítems del menú desplegable */
.dropdown-menu .dropdown-item {
  margin-bottom: 10px; /* Ajusta este valor según el espacio deseado */
}

/* Ajusta el espaciado en dispositivos móviles también */
@media (max-width: 991.98px) {
  .navbar-nav .nav-item {
    margin-bottom: 10px; /* Ajusta este valor según el espacio deseado */
  }
}
</style>
