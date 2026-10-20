import { createRouter, createWebHashHistory } from 'vue-router';
import ResultView from '../components/ResultView.vue';
import resultsData from '../generated/results.json';

export default createRouter({
  history: createWebHashHistory(process.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: () =>
        resultsData.results.length
          ? `/result/${[...resultsData.results].sort().reverse()[0]}`
          : '/result/none',
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
    {
      path: '/result/:result',
      name: 'ResultView',
      component: ResultView,
    },
  ],
});
