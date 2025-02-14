<template>
  <div class="gallery-container">
    <div class="gallery">
      <GalleryItem
        v-for="(item, index) in items"
        :key="index"
        :image="item.image"
        :title="item.title"
      />
    </div>
  </div>
</template>

<script>
import GalleryItem from "./GalleryItem.vue";

export default {
  components: {
    GalleryItem,
  },
  props: {
    items: {
      type: Array,
      required: true,
    },
  },
};
</script>

<style scoped>
.gallery-container {
  width: 70%;
  margin: 0 auto;
  padding-top: 50px;
  padding-bottom: 50px;
}

.gallery {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 20px;
}
@media (max-width: 700px) {
  .gallery {
    flex-direction: column;
  }
}
</style>
