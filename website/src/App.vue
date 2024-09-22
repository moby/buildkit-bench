<template>
  <div class="container">
    <aside class="sidebar">
      <ResultsList />
    </aside>
    <section class="result-view">
      <router-view></router-view>
    </section>
  </div>
</template>

<script>
import ResultsList from './components/ResultsList.vue';

export default {
  name: 'App',
  components: {
    ResultsList
  }
};
</script>

<style scoped>
.container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  flex: 0 0 auto;
  background-color: #f8f9fa;
  padding: 10px;
}

.result-view {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
