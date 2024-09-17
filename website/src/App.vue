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
  width: 200px;
  background-color: #f8f9fa;
  border-right: 1px solid #dee2e6;
  padding: 10px;
}

.result-view {
  flex-grow: 1;
  position: relative;
}
</style>
