<template>
  <div class="result">
    <iframe :src="resultUrl"></iframe>
  </div>
</template>

<script>
import { publicPath } from '../../vue.config';

export default {
  name: 'ResultView',
  computed: {
    resultUrl() {
      const resultLink = `${publicPath}result/${this.$route.params.result}/index.html`;
      console.log(`Loading result ${resultLink}`);
      return resultLink;
    }
  }
};
</script>

<style scoped>
.result {
  margin-left: 250px;
  height: 100vh;
  overflow: hidden;
  position: relative;
}

iframe {
  width: 100%;
  height: 100%;
  border: none;
  position: fixed;
  top: 0;
  left: 250px;
}
</style>
