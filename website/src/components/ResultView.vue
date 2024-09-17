<template>
  <div class="result">
    <iframe :src="resultUrl"></iframe>
  </div>
</template>

<script>

export default {
  name: 'ResultView',
  computed: {
    resultUrl() {
      return `${process.env.BASE_URL}result/${this.$route.params.result}/index.html`;
    }
  }
};
</script>

<style scoped>
.result {
  height: 100vh;
  overflow: hidden;
  position: relative;
}

iframe {
  width: calc(100% - 250px);
  height: 100%;
  border: none;
  position: fixed;
  top: 0;
  left: 250px;
}
</style>
