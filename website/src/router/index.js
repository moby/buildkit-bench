import { createRouter, createWebHashHistory } from 'vue-router';
import ResultView from '../components/ResultView.vue';

export default createRouter({
  history: createWebHashHistory(process.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/overview',
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
    { path: '/overview/:result?', name: 'Overview', component: ResultView },
    {
      path: '/result/:result',
      name: 'ResultView',
      component: ResultView,
    },
  ],
});
