<template>
  <div id="app">
    <ResultsList />
    <router-view></router-view>
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

<style>
</style>
