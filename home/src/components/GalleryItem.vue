<template>
  <div class="gallery-item">
    <img :src="image" alt="Gallery item image" />
    <h3>{{ title }}</h3>
  </div>
</template>

<script>
export default {
  props: {
    image: String,
    title: String,
  },
};
</script>
*//* border: 1px solid #ddd; //** */
<style scoped>
.gallery-item {
  padding: 10px;
  text-align: center;
  flex: 1 1 calc(33.3% - 20px);
  box-sizing: border-box;
}

.gallery-item img {
  max-width: 100%;
  height: auto;
}
.gallery-item h3 {
  font-size: medium;
  text-align: left;
}
</style>
