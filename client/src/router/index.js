import Vue from 'vue'
import Router from 'vue-router'
import init from '@/components/init'
import login from '@/components/login'
import register from '@/components/register'
import me from '@/components/me'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'init',
      component: init
    },
    {
      path: '/login',
      name: 'login',
      component: login
    },
    {
      path: '/register',
      name: 'register',
      component: register
    },
    {
      path: '/me',
      name: 'me',
      component: me
    }
  ]
})
