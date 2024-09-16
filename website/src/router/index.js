import {createRouter, createWebHashHistory} from 'vue-router';
import ResultView from '../components/ResultView.vue';

export default createRouter({
  history: createWebHashHistory(process.env.BASE_URL),
  routes: [
    {
      path: '/result/:result',
      name: 'ResultView',
      component: ResultView
    }
  ]
});
