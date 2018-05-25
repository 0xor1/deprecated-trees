<template>
  <div class="c-c">
    <div class="form">
      <input class="bg2 c2" v-model="email" placeholder="email" type="email"/>
      <input class="bg2 c2" v-model="pwd" placeholder="password" type="password"/>
      <button v-on:click="login">login</button>
      <button v-on:click="register">register</button>
    </div>
  </div>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  export default {
    name: 'login',
    data () {
      return {
        email: '',
        pwd: '',
        err: null
      }
    },
    methods: {
      login () {
        api.v1.centralAccount.authenticate(this.email, this.pwd).then(() => {
          router.push('/me')
        }).catch(() => {
          this.err = 'login failed'
        })
      },
      register () {
        router.push('/register')
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
