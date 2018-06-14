import Vue from 'vue'
import Router from 'vue-router'
import init from '@/components/init'
import login from '@/components/login'
import register from '@/components/register'
import confirmEmail from '@/components/confirmEmail'
import activate from '@/components/activate'
import app from '@/components/app'
import projects from '@/components/projects'
import task from '@/components/task'

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
      path: '/confirmEmail',
      name: 'confirmEmail',
      component: confirmEmail
    },
    {
      path: '/activate/:activationCode',
      name: 'activate',
      component: activate
    },
    {
      path: '/app/region/:region/shard/:shard/account/:account',
      name: 'app',
      component: app,
      children: [
        {
          path: 'projects',
          name: 'projects',
          component: projects
        },
        {
          path: 'project/:project/task/:task',
          name: 'task',
          component: task
        }
      ]
    }
  ]
})
