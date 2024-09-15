import Vue from 'vue';
import Router from 'vue-router';
import ResultView from '@/components/ResultView.vue';
import { publicPath } from '../../vue.config';

Vue.use(Router);

export default new Router({
  base: publicPath,
  routes: [
    {
      path: '/result/:result',
      name: 'ResultView',
      component: ResultView
    }
  ]
});
