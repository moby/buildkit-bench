import Vue from 'vue';
import Router from 'vue-router';
import ResultView from '@/components/ResultView.vue';

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: '/result/:result',
      name: 'ResultView',
      component: ResultView
    }
  ]
});
