// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import Vuetify from 'vuetify'
import app from '@/app'
import router from '@/router'
import 'vuetify/dist/vuetify.min.css'
import 'material-design-icons-iconfont/dist/material-design-icons.scss'
import colors from 'vuetify/es5/util/colors'

Vue.config.productionTip = false

Vue.use(Vuetify, {
  theme: {
    primary: '#9FC657',
    secondary: '#405A0F',
    accent: '#7FB21E',
    error: colors.red.base,
    warning: colors.yellow.base,
    info: colors.blue.base,
    success: colors.green.base
  }
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  template: '<app/>',
  components: { app }
})
