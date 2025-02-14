<template>
  <div class="contact-form">
    <form @submit.prevent="sendEmail" class="horizontal-form">
      <div class="form-group">
        <div class="message">
          <p>
            ¡Estamos aquí para ayudarte! Por favor, completa el formulario a la
            derecha.
          </p>
        </div>
        <div class="input-container">
          <label for="name">Nombre:</label>
          <input type="text" v-model="form.name" id="name" required />
          <label for="email">Correo Electrónico:</label>
          <input type="email" v-model="form.email" id="email" required />
          <label for="message">Mensaje:</label>
          <textarea v-model="form.message" id="message" required></textarea>
          <button type="submit" class="btn">Enviar</button>
        </div>
      </div>
    </form>
    <div v-if="sent" class="success-message">¡Mensaje enviado con éxito!</div>
    <div v-if="error" class="error-message">
      Hubo un error al enviar el mensaje.
    </div>
  </div>
</template>

<script>
import emailjs from "emailjs-com";

export default {
  data() {
    return {
      form: {
        name: "",
        email: "",
        message: "",
      },
      sent: false,
      error: false,
    };
  },
  methods: {
    sendEmail() {
      const templateParams = {
        from_name: this.form.name,
        from_email: this.form.email,
        message: this.form.message,
      };

      emailjs
        .send(
          "service_8xpzvzp", // Reemplaza con tu Service ID
          "template_67e2dnn", // Reemplaza con tu Template ID
          templateParams,
          "xCp5KRbQFSsDkp1Bv" // Reemplaza con tu User ID
        )
        .then(
          () => {
            this.sent = true;
            // Reinicia el formulario
            this.form.name = "";
            this.form.email = "";
            this.form.message = "";
          },
          (error) => {
            this.error = true;
            console.error("Error:", error); // Log de error en consola
          }
        );
    },
  },
};
</script>

<style scoped>
.contact-form {
  width: 80%;
  margin: auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 10px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  text-align: left;
  margin-top: 50px;
}

h1 {
  text-align: center;
  padding-bottom: 20px;
}

.horizontal-form {
  display: flex; /* Cambiado a flex para colocar el mensaje y el formulario en columnas */
  align-items: center; /* Centra verticalmente los elementos */
}

.form-group {
  display: flex; /* Utiliza flex para permitir la disposición horizontal */
  justify-content: space-between; /* Espaciado entre el mensaje y el formulario */
  align-items: center; /* Alinea verticalmente el contenido */
}

.message {
  width: 40%; /* Ajusta el ancho del mensaje */
  padding-right: 20px; /* Espacio entre el mensaje y el formulario */
  text-align: left; /* Centra el texto dentro del contenedor */
}

.input-container {
  width: 55%; /* Ajusta el ancho del contenedor de entrada */
}

.input-container label {
  margin-top: 10px; /* Espacio entre las etiquetas */
}

.input-container input,
.input-container textarea {
  width: 100%; /* Los campos ocuparán el 100% de ancho del contenedor */
  padding: 8px; /* Espaciado para los campos de entrada */
  border: 1px solid #ccc;
  border-radius: 5px;
}

textarea {
  resize: vertical; /* Permitir cambiar el tamaño verticalmente */
}

.btn {
  display: block;
  width: auto; /* Ajustar el ancho para el botón */
  padding: 10px 20px; /* Espaciado para el botón */
  background-color: #01695b;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-top: 10px; /* Añadido margen superior para el botón */
}

.btn:hover {
  background-color: #014d43;
}

.success-message,
.error-message {
  margin-top: 20px;
  padding: 10px;
  text-align: center;
}

.success-message {
  background-color: #d4edda;
  color: #155724;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
}
</style>
